package atlib

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	appbsky "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
)

var FeedPost_Embed appbsky.FeedPost_Embed

type Facet_Type int

const (
	Facet_Link Facet_Type = iota + 1
	Facet_Mention
	Facet_Tag
)

// construct the post
type PostBuilder struct {
	Text  string
	Facet []Facet
	Embed Embed
}

type Facet struct {
	Ftype   Facet_Type
	Value   string
	T_facet string
}

type Embed struct {
	Link           Link
	Images         []Image
	UploadedImages []lexutil.LexBlob
}

type Link struct {
	Title       string
	Uri         url.URL
	Description string
	Thumb       lexutil.LexBlob
}

type Image struct {
	Title string
	Uri   url.URL
}

// Create a simple post with text
func NewPostBuilder(text string) PostBuilder {
	return PostBuilder{
		Text:  text,
		Facet: []Facet{},
	}
}

// Create a Richtext Post with facests
func (pb PostBuilder) WithFacet(ftype Facet_Type, value string, text string) PostBuilder {

	pb.Facet = append(pb.Facet, Facet{
		Ftype:   ftype,
		Value:   value,
		T_facet: text,
	})

	return pb
}

// Create a Post with external links
func (pb PostBuilder) WithExternalLink(title string, link url.URL, description string, thumb lexutil.LexBlob) PostBuilder {

	pb.Embed.Link.Title = title
	pb.Embed.Link.Uri = link
	pb.Embed.Link.Description = description
	pb.Embed.Link.Thumb = thumb

	return pb
}

// Create a Post with images
func (pb PostBuilder) WithImages(blobs []lexutil.LexBlob, images []Image) PostBuilder {

	pb.Embed.Images = images
	pb.Embed.UploadedImages = blobs

	return pb
}

// Build the request
func (pb PostBuilder) Build() (appbsky.FeedPost, error) {

	post := appbsky.FeedPost{}

	post.Text = pb.Text
	post.LexiconTypeID = "app.bsky.feed.post"
	post.CreatedAt = time.Now().Format(time.RFC3339)

	// RichtextFacet Section
	// https://docs.bsky.app/docs/advanced-guides/post-richtext

	Facets := []*appbsky.RichtextFacet{}

	for _, f := range pb.Facet {
		facet := &appbsky.RichtextFacet{}
		features := []*appbsky.RichtextFacet_Features_Elem{}
		feature := &appbsky.RichtextFacet_Features_Elem{}

		switch f.Ftype {

		case Facet_Link:
			{
				feature = &appbsky.RichtextFacet_Features_Elem{
					RichtextFacet_Link: &appbsky.RichtextFacet_Link{
						LexiconTypeID: f.Ftype.String(),
						Uri:           f.Value,
					},
				}
			}

		case Facet_Mention:
			{
				feature = &appbsky.RichtextFacet_Features_Elem{
					RichtextFacet_Mention: &appbsky.RichtextFacet_Mention{
						LexiconTypeID: f.Ftype.String(),
						Did:           f.Value,
					},
				}
			}

		case Facet_Tag:
			{
				feature = &appbsky.RichtextFacet_Features_Elem{
					RichtextFacet_Tag: &appbsky.RichtextFacet_Tag{
						LexiconTypeID: f.Ftype.String(),
						Tag:           f.Value,
					},
				}
			}

		}

		features = append(features, feature)
		facet.Features = features

		ByteStart, ByteEnd, err := findSubstring(post.Text, f.T_facet)
		if err != nil {
			return post, fmt.Errorf("unable to find the substring: %v , %v", f.T_facet, err)
		}

		index := &appbsky.RichtextFacet_ByteSlice{
			ByteStart: int64(ByteStart),
			ByteEnd:   int64(ByteEnd),
		}
		facet.Index = index

		Facets = append(Facets, facet)
	}

	post.Facets = Facets

	// Embed Section (either external links or images)
	// As of now it allows only one Embed type per post:
	// https://github.com/bluesky-social/indigo/blob/main/api/bsky/feedpost.go
	if pb.Embed.Link != (Link{}) {

		FeedPost_Embed.EmbedExternal = &appbsky.EmbedExternal{
			LexiconTypeID: "app.bsky.embed.external",
			External: &appbsky.EmbedExternal_External{
				Title:       pb.Embed.Link.Title,
				Uri:         pb.Embed.Link.Uri.String(),
				Description: pb.Embed.Link.Description,
				Thumb:       &pb.Embed.Link.Thumb,
			},
		}

	} else {
		if len(pb.Embed.Images) != 0 && len(pb.Embed.Images) == len(pb.Embed.UploadedImages) {

			EmbedImages := appbsky.EmbedImages{
				LexiconTypeID: "app.bsky.embed.images",
				Images:        make([]*appbsky.EmbedImages_Image, len(pb.Embed.Images)),
			}

			for i, img := range pb.Embed.Images {
				EmbedImages.Images[i] = &appbsky.EmbedImages_Image{
					Alt:   img.Title,
					Image: &pb.Embed.UploadedImages[i],
				}
			}

			FeedPost_Embed.EmbedImages = &EmbedImages

		}
	}

	// avoid error when trying to marshal empty field (*bsky.FeedPost_Embed)
	if len(pb.Embed.Images) != 0 || pb.Embed.Link.Title != "" {
		post.Embed = &FeedPost_Embed
	}

	return post, nil
}

func (f Facet_Type) String() string {
	switch f {
	case Facet_Link:
		return "app.bsky.richtext.facet#link"
	case Facet_Mention:
		return "app.bsky.richtext.facet#mention"
	case Facet_Tag:
		return "app.bsky.richtext.facet#tag"
	default:
		return "Unknown"
	}
}
func findSubstring(s, substr string) (int, int, error) {
	index := strings.Index(s, substr)
	if index == -1 {
		return 0, 0, errors.New("substring not found")
	}
	return index, index + len(substr), nil
}
