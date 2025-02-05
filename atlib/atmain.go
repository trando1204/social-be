package atlib

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bluesky-social/indigo/api/atproto"
	appbsky "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
)

const defaultPDS = "https://bsky.social"

var blob []lexutil.LexBlob

// Wrapper over the atproto xrpc transport
type BskyAgent struct {
	// xrpc transport, a wrapper around http server
	client     *xrpc.Client
	handle     string
	password   string
	email      string
	inviteCode string
}

// Creates new BlueSky Agent
func NewAgent(ctx context.Context, server string, handle string, password string) BskyAgent {
	if server == "" {
		server = defaultPDS
	}
	return BskyAgent{
		client: &xrpc.Client{
			Client: new(http.Client),
			Host:   server,
		},
		handle:   handle,
		password: password,
	}
}

func (c *BskyAgent) SetEmail(email string) {
	c.email = email
}

func (c *BskyAgent) SetAdminToken(adminPass string) {
	c.client.AdminToken = &adminPass
}

func (c *BskyAgent) SetInviteCode(inviteCode string) {
	c.inviteCode = inviteCode
}

func (c *BskyAgent) Connect(ctx context.Context) error {
	// Authenticate with the Bluesky server
	input_for_session := &atproto.ServerCreateSession_Input{
		Identifier: c.handle,
		Password:   c.password,
	}
	session, err := atproto.ServerCreateSession(ctx, c.client, input_for_session)

	if err != nil {
		return fmt.Errorf("UNABLE TO CONNECT: %v", err)
	}

	// Access Token is used to make authenticated requests
	// Refresh Token allows to generate a new Access Token
	c.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJwt,
		RefreshJwt: session.RefreshJwt,
		Handle:     session.Handle,
		Did:        session.Did,
	}
	return nil
}

func (c *BskyAgent) CreateInviteCode(ctx context.Context, adminToken string) (error, string) {
	input_for_create_invite_code := &atproto.ServerCreateInviteCode_Input{
		UseCount: 1,
	}
	c.client.AdminToken = &adminToken
	accountOutput, err := atproto.ServerCreateInviteCode(ctx, c.client, input_for_create_invite_code)
	if err != nil {
		log.Printf("UNABLE TO CREATE INVITE CODE: %v", err)
		return err, ""
	}
	return nil, accountOutput.Code
}

func (c *BskyAgent) CreateAccount(ctx context.Context, email, inviteCode string) error {
	// Authenticate with the Bluesky server
	input_for_create := &atproto.ServerCreateAccount_Input{
		Handle:     c.handle,
		Password:   &c.password,
		Email:      &email,
		InviteCode: &inviteCode,
	}
	_, err := atproto.ServerCreateAccount(ctx, c.client, input_for_create)
	if err != nil {
		return fmt.Errorf("UNABLE TO CREATE ACCOUNT: %v", err)
	}
	return nil
}

func (c *BskyAgent) UploadImages(ctx context.Context, images ...Image) ([]lexutil.LexBlob, error) {
	for _, img := range images {
		getImage, err := getImageAsBuffer(img.Uri.String())
		if err != nil {
			log.Printf("Couldn't retrive the image: %v , %v", img, err)
		}

		resp, err := atproto.RepoUploadBlob(ctx, c.client, bytes.NewReader(getImage))
		if err != nil {
			return nil, err
		}

		blob = append(blob, lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: resp.Blob.MimeType,
			Size:     resp.Blob.Size,
		})
	}
	return blob, nil
}

func (c *BskyAgent) UploadImage(ctx context.Context, image Image) (*lexutil.LexBlob, error) {
	getImage, err := getImageAsBuffer(image.Uri.String())
	if err != nil {
		log.Printf("Couldn't retrive the image: %v , %v", image, err)
	}

	resp, err := atproto.RepoUploadBlob(ctx, c.client, bytes.NewReader(getImage))
	if err != nil {
		return nil, err
	}

	blob := lexutil.LexBlob{
		Ref:      resp.Blob.Ref,
		MimeType: resp.Blob.MimeType,
		Size:     resp.Blob.Size,
	}

	return &blob, nil
}

func (c *BskyAgent) PostToFeed(ctx context.Context, post appbsky.FeedPost) (string, string, error) {

	post_input := &atproto.RepoCreateRecord_Input{
		// collection: The NSID of the record collection.
		Collection: "app.bsky.feed.post",
		// repo: The handle or DID of the repo (aka, current account).
		Repo: c.client.Auth.Did,
		// record: The record itself. Must contain a $type field.
		Record: &lexutil.LexiconTypeDecoder{Val: &post},
	}

	response, err := atproto.RepoCreateRecord(ctx, c.client, post_input)
	if err != nil {
		return "", "", fmt.Errorf("unable to post, %v", err)
	}

	return response.Cid, response.Uri, nil
}

func getImageAsBuffer(imageURL string) ([]byte, error) {
	// Fetch image
	response, err := http.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Check response status
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", response.Status)
	}

	// Read response body
	imageData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
