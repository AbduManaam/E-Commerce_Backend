package cloudinaryutil

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Client struct {
	cld *cloudinary.Cloudinary
}

func New(cloudName, apiKey, apiSecret string) (*Client, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &Client{cld: cld}, nil
}
func (c *Client) UploadImage(
	ctx context.Context,
	file multipart.File,
	filename string,
	productID uint, // add this
) (string, string, error) {

	folder := fmt.Sprintf("products/%d", productID) // dynamic folder per product

	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:   folder,
		PublicID: filename, 
	})
	if err != nil {
		return "", "", err
	}

	return resp.SecureURL, resp.PublicID, nil
}


func (c *Client) DeleteImage(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	return err
}
