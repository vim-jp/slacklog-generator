package slackadapter

import "context"

// CursorIterator is requirements of IterateCursor iterates with cursor.
type CursorIterator interface {
	Iterate(context.Context, Cursor) (Cursor, error)
}

// CursorIteratorFunc is a function which implements CursorIterator.
type CursorIteratorFunc func(context.Context, Cursor) (Cursor, error)

// Iterate is an implementation for CursorIterator.
func (fn CursorIteratorFunc) Iterate(ctx context.Context, c Cursor) (Cursor, error) {
	return fn(ctx, c)
}

// IterateCursor iterates CursorIterator until returning empty cursor.
func IterateCursor(ctx context.Context, iter CursorIterator) error {
	var c Cursor
	for {
		err := ctx.Err()
		if err != nil {
			return err
		}
		next, err := iter.Iterate(ctx, c)
		if err != nil {
			return err
		}
		if next == Cursor("") {
			return nil
		}
		c = next
	}
}
