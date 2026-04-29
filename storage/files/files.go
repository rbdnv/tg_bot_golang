package files

import (
	"context"
	"encoding/gob"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"project/lib/e"
	"project/storage"
	"strconv"
	"sync"
)

type Storage struct {
	basePath string
	mu       sync.Mutex
	counts   map[int64]int
}

const defaultPerm = 0o774

func New(basePath string) *Storage {
	return &Storage{
		basePath: basePath,
		counts:   make(map[int64]int),
	}
}

func (s *Storage) Init(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return os.MkdirAll(s.basePath, defaultPerm)
}

func (s *Storage) SaveLink(ctx context.Context, userID int64, link string) (err error) {
	defer func() { err = e.WrapIfErr("save link", err) }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	page := &storage.Page{
		UserID: userID,
		URL:    link,
		Link:   link,
	}

	fPath := filepath.Join(s.basePath, userDir(userID))
	if err := os.MkdirAll(fPath, defaultPerm); err != nil {
		return err
	}

	fName, err := fileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fName)
	if _, err := os.Stat(fPath); err == nil {
		return storage.ErrDuplicateLink
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	return gob.NewEncoder(file).Encode(page)
}

func (s *Storage) GetRandomLink(ctx context.Context, userID int64) (link string, err error) {
	defer func() { err = e.WrapIfErr("get random link", err) }()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	path := filepath.Join(s.basePath, userDir(userID))
	files, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", storage.ErrNoSavedPages
	}

	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", storage.ErrNoSavedPages
	}

	file := files[rand.Intn(len(files))]
	page, err := s.decodePage(filepath.Join(path, file.Name()))
	if err != nil {
		return "", err
	}

	return firstNonEmpty(page.Link, page.URL), nil
}

func (s *Storage) IncrementUserMessages(ctx context.Context, userID int64) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[userID]++
	return s.counts[userID], nil
}

func (s *Storage) CountUserMessages(ctx context.Context, userID int64) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.counts[userID], nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("decode page", err)
	}
	defer func() { _ = f.Close() }()

	var p storage.Page
	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("decode page", err)
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}

func userDir(userID int64) string {
	return strconv.FormatInt(userID, 10)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

var _ storage.Storage = (*Storage)(nil)
