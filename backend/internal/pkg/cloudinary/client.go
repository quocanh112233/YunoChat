package cloudinary

import (
	"crypto/sha1"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Config struct {
	CloudName string
	APIKey    string
	APISecret string
}

type UploadParams struct {
	APIKey         string `json:"api_key"`
	Timestamp      int64  `json:"timestamp"`
	Signature      string `json:"signature"`
	Folder         string `json:"folder"`
	PublicID       string `json:"public_id"`
	Transformation string `json:"transformation"`
	Eager          string `json:"eager"`
}

type SignatureResponse struct {
	UploadURL          string       `json:"upload_url"`
	UploadParams       UploadParams `json:"upload_params"`
	CloudinaryPublicID string       `json:"cloudinary_public_id"`
	ExpiresAt          time.Time    `json:"expires_at"`
}

type Client interface {
	GenerateSignature(publicID string) (*SignatureResponse, error)
}

type client struct {
	config Config
}

func NewClient(cfg Config) Client {
	return &client{config: cfg}
}

func (c *client) GenerateSignature(publicID string) (*SignatureResponse, error) {
	timestamp := time.Now().Unix()
	folder := "avatars"
	transformation := "c_fill,w_256,h_256,q_auto,f_auto"
	eager := "c_fill,w_64,h_64"

	// Params to sign (không bao gồm api_key)
	params := map[string]string{
		"timestamp":      fmt.Sprintf("%d", timestamp),
		"folder":         folder,
		"public_id":      publicID,
		"transformation": transformation,
		"eager":          eager,
	}

	// 1. Sort keys
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. Build string
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	strToSign := strings.Join(pairs, "&") + c.config.APISecret

	// 3. SHA1
	hash := sha1.New()
	hash.Write([]byte(strToSign))
	signature := fmt.Sprintf("%x", hash.Sum(nil))

	fullPublicID := fmt.Sprintf("%s/%s", folder, publicID)

	resp := &SignatureResponse{
		UploadURL: fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", c.config.CloudName),
		UploadParams: UploadParams{
			APIKey:         c.config.APIKey,
			Timestamp:      timestamp,
			Signature:      signature,
			Folder:         folder,
			PublicID:       publicID,
			Transformation: transformation,
			Eager:          eager,
		},
		CloudinaryPublicID: fullPublicID,
		ExpiresAt:          time.Now().Add(5 * time.Minute),
	}

	return resp, nil
}
