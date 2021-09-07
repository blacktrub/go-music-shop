package main

import (
	"errors"
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

type Storage interface {
	Create(album) album
	Read() []album
	ReadOne(id string) (album, error)
	Update(id string, a album) (album, error)
	Delete(id string) error
}

type MemoryStorage struct {
	albums []album
}

func (s MemoryStorage) Create(am album) album {
	s.albums = append(s.albums, am)
	return am
}

func (s MemoryStorage) ReadOne(id string) (album, error) {
	for _, a := range s.albums {
		if a.ID == id {
			return a, nil
		}
	}
	return album{}, errors.New("not found")
}

func (s MemoryStorage) Read() []album {
	return s.albums
}

func (s MemoryStorage) Update(id string, newAlbum album) (album, error) {
	for i := range s.albums {
		if s.albums[i].ID == id {
			s.albums[i] = newAlbum
			return s.albums[i], nil
		}
	}
	return album{}, errors.New("not found")
}

func (s MemoryStorage) Delete(id string) error {
	for i, a := range s.albums {
		if a.ID == id {
			s.albums = append(s.albums[:i], s.albums[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func NewMemoryStorage() MemoryStorage {
	var albums = []album{
		{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
		{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
		{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
	}
	return MemoryStorage{albums: albums}
}

type PostgresStorage struct {
	db *sql.DB
}

func (p PostgresStorage) CreateSchema() error {
	_, err := p.db.Exec("create table if not exists albums (ID char(16) primary key, Title char(128), Artist char(128), Price decimal)")
	return err
}

func NewPostgresStorage() PostgresStorage {
	connStr := "user=user dbname=db password=pass sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	storage := PostgresStorage{db: db}
	err = storage.CreateSchema()
	if err != nil {
		log.Fatal(err)
	}
	return storage
}

func (p PostgresStorage) Create(am album) album {
	p.db.Exec("insert into albums(ID, Title, Artist, Price) values($1, $2, $3, $4)", am.ID, am.Title, am.Artist, am.Price)
	return am
}

func (p PostgresStorage) ReadOne(id string) (album, error) {
	var album album
	row := p.db.QueryRow("select * from albums where id = $1", id)
	err := row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return album, errors.New("not found")
		}
		return album, err
	}
	return album, nil
}

func (p PostgresStorage) Update(id string, a album) (album, error) {
	result, _ := p.db.Exec("update albums set Title=$1, Artist=$2, Price=$3 where id=$4", a.Title, a.Artist, a.Price, id)
	err := handleNotFound(result)
	return a, err
}

func (p PostgresStorage) Delete(id string) error {
	result, _ := p.db.Exec("delete from albums where id=$1", id)
	err := handleNotFound(result)
	return err
}

func handleNotFound(result sql.Result) error {
	countAffected, _ := result.RowsAffected()
	if countAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (p PostgresStorage) Read() []album {
	var albums []album
	rows, _ := p.db.Query("select * from albums")
	defer rows.Close()

	for rows.Next() {
		var a album
		rows.Scan(&a.ID, &a.Title, &a.Artist, &a.Price)
		albums = append(albums, a)
	}
	return albums
}

func NewStorage() Storage {
	return NewPostgresStorage()
}
