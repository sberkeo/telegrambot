package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"os"
	"path/filepath"	
	"math/rand"
	"time"
)

const APIURL = "https://api.micropaytoken.com/dummytelegrambot/"
const telegramBotToken = "" //token read from config.json
const helpCommand = "/help"
const referenceCodeCommand = "/refererlink"
const walletCommand = "/accountinfo"
const reflistCommand = "/referals"
const buyCommand = "/buytoken"
const sellCommand = "/selltoken"
const sendCommand = "/sendtoken"
const sponsorCommand = "/sponsor"
const sociallinksCommand = "/social"
const deleteLastCommand = "/delete_last"
const keywardlinescounter = "3,3,2,2,1"
const dummyCommand = "/dummycommand"
const createteamGroupCommand = "/team"

/*
/start		: Start or reset chat.
/help		: Help response message.
/buytoken	: Buy MicroPayToken
/selltoken	: Sell MicroPayToken
/sendtoken	: Send MicroPayToken to destination wallet
/accountinfo	: Wallet Information
/privatekey	: Privatekey Information
/refererlink	: Your Referer Link
/referercount	: Your Referer Counter
/delete		: Delete Last Message
*/

var commands = []string{helpCommand, referenceCodeCommand, walletCommand, dummyCommand, sponsorCommand, reflistCommand, createteamGroupCommand,sociallinksCommand,  buyCommand, sellCommand, sendCommand,dummyCommand,deleteLastCommand, dummyCommand,dummyCommand,dummyCommand}
var labels = []string{"Help", "Referrance Code","Wallet Info", " " , "Sponsor", "Referals", "Team", "SocialLinks","Buy", "Sell", "Send", " ", "Close",  " " , " ", " "}
var responses = []string{getHelpInfoMessage(),  "Clicked wallet Info", "Clicked to buy_token", "Clicked to sell token","Clicked to send token", "Delete last message."}

var lastMessageID int
var XSuperGuid string

// The TokenConfig struct represents the structure in the JSON file
type TokenConfig struct {
	Token string `json:"token"`
}

// User data structure (struct)
type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsPremium    bool   `json:"is_premium"`
	Sponsor	     string `json:"sponsor"`
}

// XResponseValue JSON data
type XResponseValue struct {
	ID           int    `json:"id"`
	TelegramID   int    `json:"telegram_id"`
	IsBot        int    `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsPremium    int    `json:"is_premium"`
	WalletPublic string `json:"wallet_public"`
	WalletPrivate string `json:"wallet_private"`
	Ref          string `json:"ref"`
	Sponsor	     string `json:"sponsor"`
	SolBalance   int     `json:"sol_balance"`
	TokenBalance int  `json:"token_balance"`
	SponsorUsername  string `json:"sponsor_username"`
}

// XUserId data struct
type XUserJson struct {
	ID             string    `json:"username"`
	SponsorRefGUID string `json:"sponsorRefGUID"`
}

func main() {
	// Read Config file path
	jsonFilePath := "config.json"
	rand.Seed(time.Now().UnixNano())

	// Read token
	token, err := readTokenFromJSON(jsonFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Assing token 
	telegramBotToken := token

	// Access bot
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	
	var username = ""

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 60})

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		if update.Message != nil && (update.Message.Entities != nil || update.Message.Text != "") {
			// Text link entity'leri kontrol et
			if update.Message.Entities != nil {
				for _, entity := range *update.Message.Entities {
					if entity.Type == "text_link" && entity.URL != "" {
						handleReferralFromMessage(bot, update.Message, entity.URL, update.Message.From.UserName, update.Message.From.ID)
					}
				}
			}
		
			// Direkt mesaj içinde URL kontrolü yap
			handleReferralFromMessage(bot, update.Message, update.Message.Text, update.Message.From.UserName, update.Message.From.ID)
		
			
			username = update.Message.From.UserName
		}

		var chatID int64
		var command string

		user := getUserFromUpdate(update)

		// Check user
		responseValue, err := sendUserCheckRequest(user)
		if err != nil {
			log.Printf("Error sending user check request: %v", err)
			continue
		}

		// Check message
		if update.Message != nil {
			chatID = update.Message.Chat.ID
			if update.Message.IsCommand() {
				command = update.Message.Command()
			} else {
				command = update.Message.Text
			}
		}

		// Check commands
		if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Chat.ID
			command = update.CallbackQuery.Data
			//username = update.CallbackQuery.UserName
		}

		// Read user.sponsor from json
		var readUserXSuperGUID XUserJson
		stringChatID := strconv.Itoa(int(chatID))
		filepathvalue := "user" + stringChatID + username + ".json"
		readJSON(filepathvalue, &readUserXSuperGUID)
		XSuperGuid = readUserXSuperGUID.SponsorRefGUID
		user.Sponsor = XSuperGuid

		// Main Switcher
		fmt.Printf("Received command: %s\n", command)
		fmt.Printf("Chat ID: %d\n", chatID)

		switch command {
		case "/start", "start":
			handleStartCommand(bot, chatID, update)
		case "/delete_last", "delete_last":
			handleDeleteLastCommand(bot, chatID, lastMessageID)
			//log.Printf("Last message deleted.")
		default:
			handleCommand(bot, chatID, command, responseValue)
		}
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, chatID int64, update tgbotapi.Update) {

	user := getUserFromUpdate(update)
	responseValue, err := sendUserCheckRequest(user)
		if err != nil {
			log.Printf("Error sending user check request: %v", err)
			}
	userFirstName := user.FirstName
	userLastName := user.LastName

	welcomeMessage := "Hello *" + userFirstName + " " + userLastName + "*!" + getSuperWelcomeMessage(responseValue)

	msg := tgbotapi.NewMessage(chatID, welcomeMessage)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createInlineKeyboard()
	bot.Send(msg)

	// Chat sıfırlanınca lastMessageID'yi sıfırla
	lastMessageID = 0
}

// Helper function that extracts the GUID value from the given URL
func extractGUIDFromURL(urlStr string) string {
	// URL'yi parse et
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}

	// Query parametrelerini al
	queryParams := parsedURL.Query()

	// "start" parametresini kontrol et
	startParam := queryParams.Get("start")
	if startParam != "" {
		return startParam
	}

	return ""
}

func handleDeleteLastCommand(bot *tgbotapi.BotAPI, chatID int64, lastMessageID int) {
	// Mesaj ID'si geçerli değilse, işlemi gerçekleştirme
	if lastMessageID == 0 {
		log.Printf("No valid message to delete.")
		return
	}

	// Mesajı sil
	deleteMessage := tgbotapi.NewDeleteMessage(chatID, lastMessageID)
	_, err := bot.Send(deleteMessage)
	if err != nil {
		log.Printf("Error deleting message: %v", err)
	}

	// Sadece silme işlemi yap, response dönme
}

func handleCommand(bot *tgbotapi.BotAPI, chatID int64, command string, responseValue *XResponseValue) {
	var responseText string

	if len(command) > 0 && command[0] != '/' {
			command = "/" + command
		}	

	switch command {
	case helpCommand, referenceCodeCommand, buyCommand, sellCommand, walletCommand, deleteLastCommand, sociallinksCommand:
		responseText = getResponseByCommand(command, responseValue)
	default:
		responseText = "Unknown Command: " + command
	}

	// Yanıt gönder
	msg := tgbotapi.NewMessage(chatID, responseText)
	//msg := tgbotapi.NewMessage(chatID, "Command: " + command + responseText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createInlineKeyboard()
	//msg.ParseMode = "Markdown"
	//msg.ParseMode = "MarkdownV2"

	// Mesajı gönderdikten sonra ID'sini kaydet
	sentMsg, err := bot.Send(msg)
	if err == nil {
		lastMessageID = sentMsg.MessageID
	}
}

func getResponseByCommand(command string, responseValue *XResponseValue) string {

	if len(command) > 0 && command[0] != '/' {
			command = "/" + command
		}
	switch command {
	case sociallinksCommand, "social":
		return getSocialLinkInfoMessage()
	case helpCommand,"help":
		return getHelpInfoMessage()
	case referenceCodeCommand,"refererlink":
		return getReferralInfoMessage(responseValue)
	case walletCommand,"accountinfo":
		return getWalletInfoMessage(responseValue)
	case buyCommand, sellCommand, deleteLastCommand:
		// Burada diğer komutlara özel mesajları döndürebilirsiniz.
		return "This command is currently unavailable."
	default:
		return "An unknown button was clicked. Command: " + command
	}
}


func createInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    var keywardlinescounter = [...]int{3, 4, 3, 1, 1, 1, 1, 1}
    var buttonline = 0
    

    
    for i := 0; i < len(labels); i += 4 {
        var rowButtons []tgbotapi.InlineKeyboardButton

        
        for j := i; j < i+keywardlinescounter[buttonline] && j < len(labels); j++ {
			    button := tgbotapi.NewInlineKeyboardButtonData(labels[j], commands[j])
			    rowButtons = append(rowButtons, button)
        }
        rows = append(rows, rowButtons)
	buttonline = buttonline +1
    }

    keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
    return keyboard
}


func parseKeywardLinesCounter(counterStr string) []int {
	var counters []int

	// Virgül ile ayrılmış değerleri döngüye al
	counterValues := strings.Split(counterStr, ",")
	for _, valueStr := range counterValues {
		counter, err := strconv.Atoi(valueStr)
		if err == nil {
			counters = append(counters, counter)
		}
	}

	return counters
}




func getReferralInfoMessage(xResponseValue *XResponseValue) string {
	return `
Return content
`
}


func getWalletInfoMessage(xResponseValue *XResponseValue) string {
	return `
Return content
`
}

func getSuperWelcomeMessage(xResponseValue *XResponseValue) string {
	return `
	Return content
	`

}


func getHelpInfoMessage() string {
	return `
*ChatBot Commands*
/start		: Start or reset chat.
/social		: Social Links.
/help		: Help response message.
/sponsor	: Your Sponsor Info.
/referals	: Your referals group list.
/buytoken	: Buy MicroPayToken
/selltoken	: Sell MicroPayToken
/sendtoken	: Send MicroPayToken to ...
/accountinfo	: Wallet Information
/privatekey	: Privatekey Information
/refererlink	: Your Referer Link
/referercount	: Your Referer Counter
/delete		: Delete Last Message
`
}

func getSocialLinkInfoMessage() string {
	return `
*Social Links:*

*Project Web Site*
[MicroPayToken Web Site](https://www.micropaytoken.com)

*MicroPayToken Solscan Page*
[MicroPayToken Web Site](https://solscan.io/token/H2kePwgw3WGamA4s3RntkWV84uATYmwzAqUfB4hYHpmC)
  
*Twitter*
[MicroPayToken Official Twitter](https://twitter.com/MicroPayToken)

*Discord*
[MicroPayToken Discord Group](https://discord.gg/D6E82zrT)

*Further questions? 
Join our Telegram group:* 
[MicroPayToken Telegram Group](https://t.me/MicroPayToken)
`
}


func getUserFromUpdate(update tgbotapi.Update) *User {
    var user User

    if update.Message != nil {
        user.ID = update.Message.From.ID
        user.IsBot = update.Message.From.IsBot
        user.FirstName = update.Message.From.FirstName
        user.LastName = update.Message.From.LastName
        user.Username = update.Message.From.UserName
        user.LanguageCode = update.Message.From.LanguageCode
        user.IsPremium = true // Bu bilgiyi gerçek veritabanından almanız gerekir.
        user.Sponsor = XSuperGuid
    } else if update.CallbackQuery != nil {
        user.ID = update.CallbackQuery.From.ID
        user.IsBot = update.CallbackQuery.From.IsBot
        user.FirstName = update.CallbackQuery.From.FirstName
        user.LastName = update.CallbackQuery.From.LastName
        user.Username = update.CallbackQuery.From.UserName
        user.LanguageCode = update.CallbackQuery.From.LanguageCode
        user.IsPremium = true // Bu bilgiyi gerçek veritabanından almanız gerekir.
        user.Sponsor = XSuperGuid
    } else if update.InlineQuery != nil {
        // Ek olarak InlineQuery kontrolü
        user.ID = update.InlineQuery.From.ID
        user.IsBot = update.InlineQuery.From.IsBot
        user.FirstName = update.InlineQuery.From.FirstName
        user.LastName = update.InlineQuery.From.LastName
        user.Username = update.InlineQuery.From.UserName
        user.LanguageCode = update.InlineQuery.From.LanguageCode
        user.IsPremium = true // Bu bilgiyi gerçek veritabanından almanız gerekir.
        user.Sponsor = XSuperGuid
    }

    return &user
}

func sendUserCheckRequest(user *User) (*XResponseValue, error) {
    // Generate JSON data for HTTP POST request
    requestBody, err := json.Marshal(user)
    if err != nil {
        return nil, fmt.Errorf("error marshaling user: %w", err)
    }

    // Create HTTP client with timeout
    client := &http.Client{
        Timeout: time.Second * 10,
    }

    // HTTP POST request
    resp, err := client.Post(APIURL, "application/json", bytes.NewBuffer(requestBody))
    if err != nil {
        return nil, fmt.Errorf("error sending user check request: %w", err)
    }
    defer func() {
        // Close response body and check for errors
        if cerr := resp.Body.Close(); cerr != nil {
            log.Printf("error closing response body: %v", cerr)
        }
    }()

    // Check HTTP status code
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    // Read HTTP response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }

    var xResponseValue XResponseValue
    err = json.Unmarshal(body, &xResponseValue)
    if err != nil {
        return nil, fmt.Errorf("error unmarshaling response body: %w", err)
    }

    return &xResponseValue, nil
}

func handleReferralFromMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, text string, username string, useridint int) {
    // Mesajın içeriğini kontrol et

	stringValueUserID := strconv.Itoa(useridint)

	if message.Text != "" && strings.HasPrefix(message.Text, "/start ") {
        // /start komutu içeriyorsa ve referans değeri varsa işlem yap
        refGUID := extractGUIDFromStartCommand(message.Text)

        if refGUID != "" {
            fmt.Printf("Referral GUID from /start command: %s\n", refGUID)

            // XResponseValue değerini oluştur
            XSuperGuid := refGUID
            fmt.Printf("XSuper GUID: %s mesaj\n", XSuperGuid)

            userjsondata := XUserJson{
                ID:             username,
                SponsorRefGUID: XSuperGuid,
            }
            
            writeJSON("user"+stringValueUserID+username+".json", userjsondata)
			writeJSON("user"+stringValueUserID+".json", userjsondata)

            // Kullanıcının referans değerini kullanarak diğer işlemleri gerçekleştir
            // ...

            // Mesajı yanıtla
            msg := tgbotapi.NewMessage(message.Chat.ID, "Referral link processed! Ref Code:"+XSuperGuid)
            msg.ReplyMarkup = createInlineKeyboard()
            bot.Send(msg)
        }
    }
}


// Verilen /start komutundan GUID değerini çıkartan yardımcı fonksiyon
func extractGUIDFromStartCommand(startCommand string) string {
    // /start komutundan sonraki kısmı al
    refPart := strings.TrimPrefix(startCommand, "/start ")
    return refPart
}

// readTokenFromJSON fonksiyonu, JSON dosyasından Token değerini okur ve geri döndürür
func readTokenFromJSON(filePath string) (string, error) {
	// JSON dosyasını oku
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// TokenConfig struct'ını oluştur
	var config TokenConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return "", err
	}

	// Token değerini geri döndür
	return config.Token, nil
}


// JSON'a yazma fonksiyonu
func writeJSON(filename string, data interface{}) error {
	fullPath := filepath.Join("jsondata/", filename)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	fmt.Printf("Veri başarıyla '%s' dosyasına yazıldı.\n", filename)
	return nil
}


func readJSON(filename string, data interface{}) error {
	fullPath := filepath.Join("jsondata/", filename)
	fmt.Println("\n\nDosya Yolu:\n\n", fullPath) // fullPath'i ekrana yazdır

	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, data)
	if err != nil {
		return err
	}

	fmt.Printf("Veri başarıyla '%s' dosyasından okundu.\n", filename)
	return nil
}