package economy

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
)

// ShopLayoutData defines the data for the shop listing image.
type ShopLayoutData struct {
	Items []ShopItem
}

// ShopItem represents a single item in the global shop.
type ShopItem struct {
	Name  string
	Price float64
}

// ReceiptLayoutData defines the data for the purchase receipt image.
type ReceiptLayoutData struct {
	ItemName  string
	Price     float64
	NewWallet float64
}

// HandleShop processes the /shop command.
func (s *Service) HandleShop(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	// TODO: Fetch actual dynamic prices from the pricing engine/market state
	items := []ShopItem{
		{Name: "Basic Rations", Price: 12.50},
		{Name: "Standard Toolkit", Price: 45.00},
		{Name: "Luxury Watch", Price: 350.00},
	}

	data := ShopLayoutData{Items: items}
	imgBytes, err := s.Composer.Compose("shop", data)
	if err != nil {
		return fmt.Errorf("failed to compose shop image: %w", err)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Global Market Shopfront",
					Description: "Prices are dynamic and change based on global supply and demand.",
					Color:       0xFFD700,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://shop.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Price changes daily"},
				},
			},
			Files: []*discordgo.File{
				{Name: "shop.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
		},
	})
}

// HandleBuy processes the /buy command.
func (s *Service) HandleBuy(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	playerIDStr := ctx.Value(commands.PlayerIDKey).(string)

	itemName, ok := getStringOption(i, "item")
	if !ok {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify an item to purchase.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// TODO: Fetch actual price from pricing engine based on itemName
	price := 45.00
	switch itemName {
	case "Basic Rations":
		price = 12.50
	case "Luxury Watch":
		price = 350.00
	}

	player, err := s.Queries.GetPlayerByDiscordID(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	walletID, err := s.getAccountByType(ctx, player.ID, db.AccountTypeWALLET)
	if err != nil {
		return fmt.Errorf("failed to fetch wallet: %w", err)
	}

	treasury, err := s.Queries.GetTreasuryAccount(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch treasury: %w", err)
	}

	_, err = s.Ledger.PostTransaction(ctx, economy.PostTransactionParams{
		SourceAccountID: walletID,
		DestAccountID:   treasury.ID,
		Amount:          price,
		Type:            db.TransactionTypeMARKETPURCHASE,
		Description:     fmt.Sprintf("Shop Purchase: %s", itemName),
	})
	if err != nil {
		if err == economy.ErrInsufficientFunds {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough funds to purchase this item.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transaction failed: %w", err)
	}

	// TODO: Add item to player inventory (JSONB update via business/player inventory logic)

	newWalletBal, _ := s.getBalance(ctx, walletID)

	data := ReceiptLayoutData{
		ItemName:  itemName,
		Price:     price,
		NewWallet: newWalletBal,
	}

	imgBytes, err := s.Composer.Compose("receipt", data)
	if err != nil {
		return fmt.Errorf("failed to compose receipt image: %w", err)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Purchase Receipt",
					Description: fmt.Sprintf("Successfully purchased %s.", itemName),
					Color:       0xFFD700,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://receipt.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Thank you for your purchase"},
				},
			},
			Files: []*discordgo.File{
				{Name: "receipt.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
