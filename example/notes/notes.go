package notes

import (
	"iter"
	"math/rand/v2"
	"slices"
)

type Note struct {
	ID      int
	Content string
}

type DB struct {
	list []Note
}

func (n *DB) All() iter.Seq[Note] {
	return func(yield func(Note) bool) {
		for _, note := range n.list {
			if !yield(note) {
				return
			}
		}
	}
}

func (n *DB) Len() int {
	return len(n.list)
}

func (n *DB) Create(content string) int {
	id := rand.Int()
	n.list = append(n.list, Note{ID: id, Content: content})
	return id
}

func (n *DB) Delete(id int) {
	n.list = slices.DeleteFunc(n.list, func(note Note) bool {
		return note.ID == id
	})
}

func (n *DB) Get(id int) (Note, bool) {
	idx := slices.IndexFunc(n.list, func(note Note) bool {
		return note.ID == id
	})

	if idx < 0 {
		return Note{}, false
	}

	return n.list[idx], true
}
