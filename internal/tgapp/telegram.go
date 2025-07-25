package tgapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Config struct {
	AppID        int
	AppHash      string
	PhoneNumber  string
	DownloadPath string
}

type TelegramDownloader struct {
	config Config
	client *telegram.Client
}

func NewTelegramDownloader(cfg Config) (*TelegramDownloader, error) {
	if err := os.MkdirAll(cfg.DownloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	client := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: "session.json"},
		DialTimeout:    30 * time.Second,
		RetryInterval:  5 * time.Second,
		MaxRetries:     3,
	})

	return &TelegramDownloader{
		config: cfg,
		client: client,
	}, nil
}

func (td *TelegramDownloader) Connect(ctx context.Context) error {
	return td.client.Run(ctx, func(ctx context.Context) error {
		if err := td.auth(ctx); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}

		// Keep the client running by not returning here
		// The download will be called from within this context
		return nil
	})
}

func (td *TelegramDownloader) auth(ctx context.Context) error {
	var password string

	client := auth.NewClient(td.client.API(), os.Stdin, td.config.AppID, td.config.AppHash)

	status, err := client.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth status: %w", err)
	}
	if !status.Authorized {
		fmt.Print("Enter your Telegram password (if set, leave empty if not): ")
		fmt.Scanln(&password)
	}

	flow := auth.NewFlow(
		auth.Constant(td.config.PhoneNumber, password,
			auth.CodeAuthenticatorFunc(func(ctx context.Context, _ *tg.AuthSentCode) (string, error) {
				fmt.Print("Enter the code you received: ")
				var code string
				fmt.Scanln(&code)
				return code, nil
			})),
		auth.SendCodeOptions{},
	)

	err = client.IfNecessary(ctx, flow)
	if err != nil {
		// If it's a password error, provide guidance
		if err.Error() == "auth flow: sign in with password: invalid password" {
			fmt.Println("2FA is enabled. Please temporarily disable 2FA in Telegram Settings → Privacy and Security → Two-Step Verification")
			fmt.Println("Or try using a different account without 2FA.")
		}
	}

	return err
}

func (td *TelegramDownloader) DownloadChannelMusic(ctx context.Context, channelUsername string, limit int) error {
	log.Printf("Starting download from channel: @%s", channelUsername)

	// Run the client and perform download within the same context
	return td.client.Run(ctx, func(ctx context.Context) error {
		if err := td.auth(ctx); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}

		api := td.client.API()

		// Get channel info
		peer, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: channelUsername,
		})
		if err != nil {
			return fmt.Errorf("failed to resolve channel: %w", err)
		}

		// Convert to input peer
		var inputPeer tg.InputPeerClass
		if len(peer.Users) > 0 {
			if user, ok := peer.Users[0].(*tg.User); ok {
				inputPeer = &tg.InputPeerUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				}
			}
		}
		if len(peer.Chats) > 0 {
			if chat, ok := peer.Chats[0].(*tg.Channel); ok {
				inputPeer = &tg.InputPeerChannel{
					ChannelID:  chat.ID,
					AccessHash: chat.AccessHash,
				}
			}
		}

		if inputPeer == nil {
			return fmt.Errorf("failed to get input peer")
		}

		// Get channel messages
		messages, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:      inputPeer,
			Limit:     limit,
			AddOffset: 0,
			OffsetID:  0,
			MaxID:     0,
			MinID:     0,
			Hash:      0,
		})
		if err != nil {
			return fmt.Errorf("failed to get messages: %w", err)
		}

		downloadedCount := 0

		// Handle different message types
		switch m := messages.(type) {
		case *tg.MessagesMessages:
			for _, msg := range m.Messages {
				if err := td.processMessage(ctx, msg); err != nil {
					log.Printf("Error processing message: %v", err)
				} else {
					downloadedCount++
				}

				if limit > 0 && downloadedCount >= limit {
					break
				}
			}
		case *tg.MessagesChannelMessages:
			for _, msg := range m.Messages {
				if err := td.processMessage(ctx, msg); err != nil {
					log.Printf("Error processing message: %v", err)
				} else {
					downloadedCount++
				}

				if limit > 0 && downloadedCount >= limit {
					break
				}
			}
		}

		log.Printf("Downloaded %d audio files", downloadedCount)
		return nil
	})
}

func (td *TelegramDownloader) processMessage(ctx context.Context, msg tg.MessageClass) error {
	message, ok := msg.(*tg.Message)
	if !ok {
		return nil
	}

	var media tg.MessageMediaClass
	if message.Media != nil {
		media = message.Media
	} else {
		return nil
	}

	var fileName string
	var fileSize int64
	var file tg.InputFileLocationClass

	switch m := media.(type) {
	case *tg.MessageMediaDocument:
		if m.Document == nil {
			return nil
		}
		doc, ok := m.Document.(*tg.Document)
		if !ok {
			return nil
		}

		// Get filename from attributes
		for _, attr := range doc.Attributes {
			if filenameAttr, ok := attr.(*tg.DocumentAttributeFilename); ok {
				fileName = filenameAttr.FileName
				break
			}
		}

		fileSize = doc.Size
		file = &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
			ThumbSize:     "",
		}

	default:
		return nil
	}

	if fileName == "" {
		fileName = fmt.Sprintf("file_%d", message.ID)
	}

	return td.downloadFile(ctx, file, fileName, fileSize)
}

func (td *TelegramDownloader) downloadFile(ctx context.Context, file tg.InputFileLocationClass, fileName string, fileSize int64) error {
	filePath := filepath.Join(td.config.DownloadPath, fileName)

	if _, err := os.Stat(filePath); err == nil {
		log.Printf("File already exists: %s", fileName)
		return nil
	}

	log.Printf("Downloading: %s (%s)", fileName, formatFileSize(fileSize))

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	api := td.client.API()

	// Download file in chunks
	offset := int64(0)
	chunkSize := int64(512 * 1024) // 512KB chunks

	for offset < fileSize {
		chunk, err := api.UploadGetFile(ctx, &tg.UploadGetFileRequest{
			Precise:      false,
			CDNSupported: false,
			Location:     file,
			Offset:       offset,
			Limit:        int(chunkSize),
		})
		if err != nil {
			return fmt.Errorf("failed to download chunk: %w", err)
		}

		uploadFile, ok := chunk.(*tg.UploadFile)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if _, err := out.Write(uploadFile.Bytes); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}

		offset += int64(len(uploadFile.Bytes))
		if int64(len(uploadFile.Bytes)) < chunkSize {
			break
		}
	}

	log.Printf("Successfully downloaded: %s", fileName)
	return nil
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
}
