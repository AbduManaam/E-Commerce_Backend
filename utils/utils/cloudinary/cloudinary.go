package cloudinaryutil

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"

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

// ✅ FIXED: Remove duplicate folder creation
func (c *Client) UploadImage(
	ctx context.Context,
	file multipart.File,
	filename string,  // Should be just the filename like "35" or "image1"
	productID uint,
) (string, string, error) {

	// Create folder path
	folder := fmt.Sprintf("products/%d", productID) // "products/35"

	// Use productID as filename if filename is empty or looks like a path
	if filename == "" || filename == folder {
		filename = strconv.Itoa(int(productID)) // Just "35"
	}

	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:   folder,     // "products/35"
		PublicID: filename,   // "35"
	})
	if err != nil {
		return "", "", err
	}

	// Result: products/35/35.jpg ✅
	return resp.SecureURL, resp.PublicID, nil
}

func (c *Client) DeleteImage(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	return err
}