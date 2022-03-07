package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/term"
)

var (
	username string
	avatar   string

	ALLOWED_AVATAR_MIME_TYPES = []string{
		"image/png",
		"image/jpeg",
	}
)

type ModifyCurrentUser struct {
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

func init() {
	flag.StringVar(&username, "username", "", "the new username for the bot")
	flag.StringVar(&avatar, "avatar", "", "the new avatar for the bot. either a `url or file path`")

	flag.Parse()

	if username == "" && avatar == "" {
		fmt.Fprintln(os.Stderr, "incorrect usage: must provide at least one of `-avatar`, `-username`")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	if username != "" {
		if l := len(username); l <= 2 || l >= 32 {
			fmt.Fprintln(os.Stderr, "username must be 2-32 characters in length.")
			os.Exit(1)
		}
	}
}

func main() {
	tkn := getToken()

	if tkn == "" {
		fmt.Fprintln(os.Stderr, "no token was provided")
		os.Exit(1)
	}

	patch := ModifyCurrentUser{
		Username: username,
	}

	if avatar != "" {
		b64, mime, err := getAvatarBase64(avatar)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading avatar: %v\n", err)
			os.Exit(1)
		} else {
			patch.Avatar = fmt.Sprintf("data:%v;base64,%v", mime, b64)
		}
	}

	if patch.Avatar == "" && patch.Username == "" {
		fmt.Fprintln(os.Stderr, "incorrect usage: must provide at least one of `-avatar`, `-username`")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	newUser, err := modifyUser(tkn, patch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to modify user: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Updated user")

	if patch.Username != "" {
		fmt.Printf("  => Username: %v#%v\n", newUser.Username, newUser.Discriminator)
	}

	if patch.Avatar != "" {
		fmt.Printf("  => Avatar  : https://cdn.discordapp.com/avatars/%v/%v.png\n", newUser.ID, newUser.Avatar)
	}
}

func getToken() string {
	token, exists := os.LookupEnv("DISCORD_TOKEN")
	if exists {
		return token
	}

	if !isTTY() {
		return ""
	}

	fmt.Print("Bot Token (imput feedback is NOT shown) >> ")
	tkn, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return ""
	}

	return string(tkn)
}

func isTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func getAvatarBase64(avatar string) (string, string, error) {
	var img []byte

	u, err := url.ParseRequestURI(avatar)
	if err == nil && u.Scheme != "" && u.Host != "" {
		img, err = getImage(avatar)
	} else {
		_, err = os.Stat(avatar)
		if err != nil {
			return "", "", err
		}

		img, err = ioutil.ReadFile(avatar)
		if err != nil {
			return "", "", err
		}
	}

	if img == nil || len(img) == 0 {
		return "", "", errors.New("image file was empty")
	}

	contentType := http.DetectContentType(img)

	if !isMimeTypeAllowed(contentType) {
		return "", "", errors.New(fmt.Sprintf("mime type %v not allowed", contentType))
	}

	encoded := base64.StdEncoding.EncodeToString(img)

	return encoded, contentType, nil
}

func isMimeTypeAllowed(mimeType string) bool {
	for _, mt := range ALLOWED_AVATAR_MIME_TYPES {
		if mt == mimeType {
			return true
		}
	}

	return false
}

func getImage(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get image from server: %v", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("image server responded with non-200 error: %v", res.Status)
	}

	bdy, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return bdy, nil
}

func modifyUser(token string, pl ModifyCurrentUser) (*User, error) {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(pl)
	if err != nil {
		return nil, errors.New("failed to serialize request payload")
	}

	req, err := http.NewRequest("PATCH", "https://discord.com/api/v10/users/@me", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to modify user: %v", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("discord responded with non-200 status: %v", res.Status)
	}

	var modifiedUser User
	err = json.NewDecoder(res.Body).Decode(&modifiedUser)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response from discord: %v", err)
	}

	return &modifiedUser, nil
}
