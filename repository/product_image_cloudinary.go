package repository

import (
	"backend/internal/domain"
	"context"
	"database/sql"
)

type ProductImageRepo interface {
    Create(ctx context.Context, img *domain.ProductImage) error
}

type productImageRepo struct {
    db *sql.DB
}

func NewProductImageRepo(db *sql.DB) ProductImageRepo {
    return &productImageRepo{db: db}
}

func (r *productImageRepo) Create(ctx context.Context, img *domain.ProductImage) error {
    query := `
        INSERT INTO product_images (product_id, url, public_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    _, err := r.db.ExecContext(ctx, query,
        img.ProductID,
        img.URL,
        img.PublicID,
        img.CreatedAt,
        img.UpdatedAt,
    )
    return err
}
