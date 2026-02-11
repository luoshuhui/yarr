package storage

import (
	"log"
)

type Folder struct {
	Id         int64  `json:"id"`
	ParentId   *int64 `json:"parent_id"`
	Title      string `json:"title"`
	IsExpanded bool   `json:"is_expanded"`
}

func (s *Storage) CreateFolder(title string, parentId *int64) *Folder {
	expanded := true
	row := s.db.QueryRow(`
		insert into folders (title, parent_id, is_expanded) values (?, ?, ?)
		on conflict (title) do update set title = ?, parent_id = ?
        returning id`,
		title, parentId, expanded,
		title, parentId,
	)
	var id int64
	err := row.Scan(&id)

	if err != nil {
		log.Print(err)
		return nil
	}
	return &Folder{Id: id, ParentId: parentId, Title: title, IsExpanded: expanded}
}

func (s *Storage) DeleteFolder(folderId int64) bool {
	_, err := s.db.Exec(`delete from folders where id = ?`, folderId)
	if err != nil {
		log.Print(err)
	}
	return err == nil
}

func (s *Storage) RenameFolder(folderId int64, newTitle string) bool {
	_, err := s.db.Exec(`update folders set title = ? where id = ?`, newTitle, folderId)
	return err == nil
}

func (s *Storage) ToggleFolderExpanded(folderId int64, isExpanded bool) bool {
	_, err := s.db.Exec(`update folders set is_expanded = ? where id = ?`, isExpanded, folderId)
	return err == nil
}

func (s *Storage) UpdateFolderParent(folderId int64, parentId *int64) bool {
	_, err := s.db.Exec(`update folders set parent_id = ? where id = ?`, parentId, folderId)
	return err == nil
}

func (s *Storage) ReorderFolders(ids []int64) {
	tx, _ := s.db.Begin()
	for i, id := range ids {
		tx.Exec(`update folders set sort_order = ? where id = ?`, i, id)
	}
	tx.Commit()
}

func (s *Storage) ListFolders() []Folder {
	result := make([]Folder, 0)
	rows, err := s.db.Query(`
		select id, parent_id, title, is_expanded
		from folders
		order by sort_order asc, title collate nocase
	`)
	if err != nil {
		log.Print(err)
		return result
	}
	for rows.Next() {
		var f Folder
		err = rows.Scan(&f.Id, &f.ParentId, &f.Title, &f.IsExpanded)
		if err != nil {
			log.Print(err)
			return result
		}
		result = append(result, f)
	}
	return result
}
