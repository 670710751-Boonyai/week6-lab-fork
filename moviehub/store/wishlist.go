package store

import (
    "context"
    "errors"
    "time"

    "github.com/jackc/pgx/v5/pgconn"
)

type WishlistItem struct {
    MovieID int64     `json:"movieId"`
    Title   string    `json:"title,omitempty"`
    TitleTh *string   `json:"titleTh,omitempty"`
    Note    *string   `json:"note"`
    AddedAt time.Time `json:"addedAt"`
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
    rows, err := s.Pool.Query(ctx,
        `SELECT w.movie_id, m.title, m.title_th, w.note, w.added_at
         FROM wishlist_items w
         JOIN movies m ON m.id = w.movie_id
         WHERE w.member_id = $1
         ORDER BY w.added_at DESC, w.movie_id DESC`, memberID)
    if err != nil { return nil, err }
    defer rows.Close()
    out := []WishlistItem{}
    for rows.Next() {
        var item WishlistItem
        if err := rows.Scan(&item.MovieID, &item.Title, &item.TitleTh, &item.Note, &item.AddedAt); err != nil {
            return nil, err
        }
        out = append(out, item)
    }
    return out, rows.Err()
}

func (s *Store) PutWishlist(ctx context.Context, memberID, movieID int64, note *string) (WishlistItem, error) {
    item := WishlistItem{MovieID: movieID}
    err := s.Pool.QueryRow(ctx,
        `INSERT INTO wishlist_items (member_id, movie_id, note)
         VALUES ($1, $2, $3)
         ON CONFLICT (member_id, movie_id) DO UPDATE SET note = EXCLUDED.note
         RETURNING note, added_at`, memberID, movieID, note).
        Scan(&item.Note, &item.AddedAt)
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23503" {
        return WishlistItem{}, ErrNoMovie
    }
    return item, err
}

func (s *Store) DeleteWishlist(ctx context.Context, memberID, movieID int64) error {
    _, err := s.Pool.Exec(ctx,
        `DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2`,
        memberID, movieID)
    return err
}
