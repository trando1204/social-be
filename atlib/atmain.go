package atlib

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"socialat/be/utils"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
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

func NewBasicAgent(ctx context.Context, server string) BskyAgent {
	return NewAgent(ctx, server, "", "")
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

func (c *BskyAgent) SetClientAuth(accessJwt, refreshJwt, handle, did string) {
	c.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
		Handle:     handle,
		Did:        did,
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

func ConnectToGetSession(ctx context.Context, server, handle, password string) (*xrpc.AuthInfo, error) {
	agent := NewAgent(ctx, server, handle, password)
	authRes, err := agent.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return authRes, nil
}

func (c *BskyAgent) Connect(ctx context.Context) (*xrpc.AuthInfo, error) {
	// Authenticate with the Bluesky server
	input_for_session := &atproto.ServerCreateSession_Input{
		Identifier: c.handle,
		Password:   c.password,
	}
	session, err := atproto.ServerCreateSession(ctx, c.client, input_for_session)

	if err != nil {
		return nil, fmt.Errorf("UNABLE TO CONNECT: %v", err)
	}

	// Access Token is used to make authenticated requests
	// Refresh Token allows to generate a new Access Token
	c.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJwt,
		RefreshJwt: session.RefreshJwt,
		Handle:     session.Handle,
		Did:        session.Did,
	}
	return c.client.Auth, nil
}

func HandlerValidSession(ctx context.Context, server string, authInfo *xrpc.AuthInfo) error {
	agent := NewBasicAgent(ctx, server)
	agent.client.Auth = authInfo
	_, err := atproto.ServerGetSession(ctx, agent.client)
	if err != nil {
		log.Printf("Get pds session failed. %v", err)
		return err
	}
	return nil
}

func CreateInviteCode(ctx context.Context, server, adminToken string) (string, error) {
	agent := NewBasicAgent(ctx, server)
	input_for_create_invite_code := &atproto.ServerCreateInviteCode_Input{
		UseCount: 1,
	}
	agent.client.AdminToken = &adminToken
	accountOutput, err := atproto.ServerCreateInviteCode(ctx, agent.client, input_for_create_invite_code)
	if err != nil {
		log.Printf("UNABLE TO CREATE INVITE CODE: %v", err)
		return "", err
	}
	return accountOutput.Code, nil
}

func CreateAccount(ctx context.Context, server, username, password, email, inviteCode string) (*atproto.ServerCreateAccount_Output, error) {
	agent := NewBasicAgent(ctx, server)
	// Get handle from server and username
	// TODO: check handle exist on pds server
	handleStr := utils.GetHandleFromUsername(server, username)
	// Authenticate with the Bluesky server
	input_for_create := &atproto.ServerCreateAccount_Input{
		Handle:     handleStr,
		Password:   &password,
		Email:      &email,
		InviteCode: &inviteCode,
	}
	out, err := atproto.ServerCreateAccount(ctx, agent.client, input_for_create)
	if err != nil {
		return nil, fmt.Errorf("UNABLE TO CREATE ACCOUNT: %v", err)
	}
	return out, nil
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

func (c *BskyAgent) GetTimeline(ctx context.Context, cursor string, limit int64) (*bsky.FeedGetTimeline_Output, error) {
	response, err := bsky.FeedGetTimeline(ctx, c.client, "reverse-chronological", cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("unable to get timeline, %v", err)
	}
	return response, nil
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
