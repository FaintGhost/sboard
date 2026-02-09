package singboxcli

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	stdjson "encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	sbjson "github.com/sagernet/sing/common/json"
)

var ErrInvalidGenerateKind = errors.New("invalid sing-box generate command")

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Format(_ context.Context, config string) (string, error) {
	text := strings.TrimSpace(config)
	if text == "" {
		return "", errors.New("config is empty")
	}

	var data any
	if err := stdjson.Unmarshal([]byte(text), &data); err != nil {
		return "", errors.New("invalid json")
	}

	var out bytes.Buffer
	encoder := sbjson.NewEncoder(&out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func (s *Service) Check(ctx context.Context, config string) (string, error) {
	text := strings.TrimSpace(config)
	if text == "" {
		return "", errors.New("config is empty")
	}

	singCtx := include.Context(ctx)

	var options option.Options
	if err := sbjson.UnmarshalContext(singCtx, []byte(text), &options); err != nil {
		return "", err
	}

	runtimeCtx, cancel := context.WithCancel(singCtx)
	defer cancel()

	instance, err := box.New(box.Options{
		Context: runtimeCtx,
		Options: options,
	})
	if err != nil {
		return "", err
	}
	instance.Close()

	return "ok", nil
}

func (s *Service) Generate(_ context.Context, kind string) (string, error) {
	command := strings.TrimSpace(kind)
	if strings.HasPrefix(command, "rand-base64-") {
		lengthText := strings.TrimPrefix(command, "rand-base64-")
		length, err := strconv.Atoi(lengthText)
		if err != nil || length <= 0 {
			return "", ErrInvalidGenerateKind
		}
		return generateRandomBase64(length)
	}

	switch command {
	case "uuid":
		return uuid.NewString(), nil

	case "reality-keypair":
		privateKey, publicKey, err := generateX25519KeyPair()
		if err != nil {
			return "", err
		}
		return "PrivateKey: " + base64.RawURLEncoding.EncodeToString(privateKey) + "\n" +
			"PublicKey: " + base64.RawURLEncoding.EncodeToString(publicKey), nil

	case "wg-keypair":
		privateKey, publicKey, err := generateX25519KeyPair()
		if err != nil {
			return "", err
		}
		return "PrivateKey: " + base64.StdEncoding.EncodeToString(privateKey) + "\n" +
			"PublicKey: " + base64.StdEncoding.EncodeToString(publicKey), nil

	case "vapid-keypair":
		privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
		if err != nil {
			return "", err
		}
		publicKey := privateKey.PublicKey()
		return "PrivateKey: " + base64.RawURLEncoding.EncodeToString(privateKey.Bytes()) + "\n" +
			"PublicKey: " + base64.RawURLEncoding.EncodeToString(publicKey.Bytes()), nil

	default:
		return "", ErrInvalidGenerateKind
	}
}

func generateRandomBase64(length int) (string, error) {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

func generateX25519KeyPair() ([]byte, []byte, error) {
	privateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	publicKey := privateKey.PublicKey()
	return privateKey.Bytes(), publicKey.Bytes(), nil
}
