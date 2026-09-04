package economy

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// ShopLayoutData defines the data for the shop listing image.
// STRICT RULE: No float64 is used for financial values.
type ShopLayoutData struct {
	Items []ShopItem
}

// ShopItem represents a single item in the global shop.
type ShopItem struct {
	Name  string
	Price decimal.Decimal
}

// ReceiptLayoutData defines the data for the purchase receipt image.
type ReceiptLayoutData struct {
	ItemName  string
	Price     decimal.Decimal
	NewWallet decimal.Decimal
}

// HandleShop processes the /shop command.
func (s *Service) HandleShop(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")
	log.Info().Msg("shop command initiated")

	// Fetch current dynamic prices for core shop items
	items := []string{"Basic Rations", "Standard Toolkit", "Luxury Watch"}
	shopItems := make([]ShopItem, 0, len(items))

	for _, itemName := range items {
		priceNumeric, err := s.Queries.GetItemCurrentPrice(ctx, itemName)
		var price decimal.Decimal
		if err != nil {
			log.Warn().Err(fmt.Errorf("failed to fetch price for %s: %w", itemName, err)).Msg("falling back to default base price")
			// Fallback base prices if DB query fails or table is empty (Layer 1 safety)
			switch itemName {
			case "Basic Rations":
				price = decimal.NewFromFloat(12.50)
			case "Standard Toolkit":
				price = decimal.NewFromFloat(45.00)
			case "Luxury Watch":
				price = decimal.NewFromFloat(350.00)
			}
		} else {
			price = numericToDecimal(priceNumeric)
		}

		shopItems = append(shopItems, ShopItem{
			Name:  itemName,
			Price: price,
		})
	}

	data := ShopLayoutData{Items: shopItems}
	imgBytes, err := s.Composer.Compose("shop", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose shop image: %w", err)).Msg("image composition failed")
		return s.sendShopTextFallback(ctx, sess, i, shopItems)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Global Market Shopfront",
					Description: "Prices are dynamic and change based on global supply and demand.",
					Color:       0xFFD700, // Gold for economy
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://shop.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Prices update at daily settlement"},
				},
			},
			Files: []*discordgo.File{
				{Name: "shop.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
		return fmt.Errorf("failed to respond to interaction: %w", err)
	}

	return nil
}

// HandleBuy processes the /buy command.
func (s *Service) HandleBuy(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in buy handler")
		return fmt.Errorf("player_id missing from context")
	}

	// 1. Parse Discord Interaction Options
	var itemName string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "item" {
			itemName = opt.StringValue()
			break
		}
	}

	if itemName == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify an item to purchase.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 2. Fetch Current Dynamic Price
	priceNumeric, err := s.Queries.GetItemCurrentPrice(ctx, itemName)
	var price decimal.Decimal
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to fetch price for %s: %w", itemName, err)).Msg("falling back to default base price")
		switch itemName {
		case "Basic Rations":
			price = decimal.NewFromFloat(12.50)
		case "Standard Toolkit":
			price = decimal.NewFromFloat(45.00)
		case "Luxury Watch":
			price = decimal.NewFromFloat(350.00)
		default:
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This item is not available in the shop.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	} else {
		price = numericToDecimal(priceNumeric)
	}

	// 3. Fetch Player and Wallet
	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("buy command failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	wallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    player.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch wallet: %w", err)).Str("player_id", playerIDStr).Msg("buy command failed")
		return fmt.Errorf("failed to fetch wallet: %w", err)
	}

	treasury, err := s.Queries.GetTreasuryAccount(ctx)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch treasury: %w", err)).Msg("buy command failed")
		return fmt.Errorf("failed to fetch treasury: %w", err)
	}

	// 4. Execute Atomic Double-Entry Ledger Transfer
	err = s.Ledger.Transfer(ctx, wallet.ID, treasury.ID, price, "MARKET_PURCHASE", player.ID, fmt.Sprintf("Shop Purchase: %s", itemName))
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("You do not have enough funds to purchase **%s** (Cost: ⊄%s).", itemName, price.String()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		log.Error().Err(fmt.Errorf("ledger transaction failed: %w", err)).Str("player_id", playerIDStr).Msg("buy command failed")
		return fmt.Errorf("ledger transaction failed: %w", err)
	}

	// 5. Fetch New Balance
	newWalletBal, _ := s.getAccountMetrics(ctx, wallet.ID)

	// 6. Compose Receipt Image
	data := ReceiptLayoutData{
		ItemName:  itemName,
		Price:     price,
		NewWallet: newWalletBal,
	}

	imgBytes, err := s.Composer.Compose("receipt", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose receipt image: %w", err)).Msg("image composition failed")
		return s.sendReceiptTextFallback(ctx, sess, i, itemName, price, newWalletBal)
	}

	// 7. Respond to Discord
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Purchase Receipt",
					Description: fmt.Sprintf("Successfully purchased **%s**.", itemName),
					Color:       0x4CAF76, // Green for successful transaction
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

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
		return fmt.Errorf("failed to respond to interaction: %w", err)
	}

	log.Info().
		Str("player_id", playerIDStr).
		Str("item", itemName).
		Str("price", price.String()).
		Msg("buy command completed successfully")

	return nil
}

// sendShopTextFallback provides a graceful degradation if the imaging compositor fails,
// ensuring the player still receives the shop listing without a bot crash.
func (s *Service) sendShopTextFallback(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, items []ShopItem) error {
	var itemList string
	for _, item := range items {
		itemList += fmt.Sprintf("• **%s**: `⊄%s`\n", item.Name, item.Price.String())
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Global Market Shopfront",
					Description: fmt.Sprintf("Prices are dynamic and change based on global supply and demand.\n\n%s", itemList),
					Color:       0xFFD700,
					Footer:      &discordgo.MessageEmbedFooter{Text: "Prices update at daily settlement"},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

// sendReceiptTextFallback provides a graceful degradation if the imaging compositor fails,
// ensuring the player still receives their purchase confirmation without a bot crash.
func (s *Service) sendReceiptTextFallback(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, itemName string, price, newWallet decimal.Decimal) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Purchase Receipt",
					Description: fmt.Sprintf("Successfully purchased **%s**.\n\nCost: `⊄%s`\nNew Wallet Balance: `⊄%s`\n*(Image generation temporarily unavailable)*", itemName, price.String(), newWallet.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
