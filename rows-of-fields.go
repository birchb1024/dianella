package dianella

import (
	"encoding/csv"
	"fmt"
	"os"
)

type RowsOfFields [][]string

func (s *Step) ReadCSV(filename string) (*Step, RowsOfFields) {
	if s.IsFailed() {
		return s, nil
	}
	f, err := os.Open(filename)
	if err != nil {
		s.Fail(err.Error())
		return s, nil
	}
	defer func() { _ = f.Close() }()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		s.Fail(err.Error())
		return s, nil
	}
	return s, records
}

func (rows RowsOfFields) SelectColumnDistinctValues(columnName string) ([]string, error) {
	var values = map[string]bool{}
	if len(rows) < 2 {
		return nil, fmt.Errorf("expected header and one row")
	}
	// Find the column number from the header row
	index := -1
	for i, field := range rows[0] {
		if field == columnName {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("could not find %s in columns %v", columnName, rows)
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		// Skip empty cells
		if len(row)-1 < index || row[index] == "" {
			continue
		}
		values[row[index]] = true
	}
	return keysOfMap(values), nil
}
func keysOfMap(m map[string]bool) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
func (rows RowsOfFields) Rows2Maps(keys []string) (map[string]map[string]string, error) {
	result := map[string]map[string]string{}
	if len(rows) < 2 {
		return nil, fmt.Errorf("expected header and one row %v", rows)
	}
	header := rows[0]
	data := rows[1:]
	for _, rec := range data {
		object := map[string]string{}
		for idx, h := range header {
			if idx >= len(rec) {
				return nil, fmt.Errorf("row too short for column %s: %v", h, rec)
			}
			object[h] = rec[idx]
		}
		pk := ""
		for _, key := range keys {
			v, ok := object[key]
			if !ok {
				return nil, fmt.Errorf("missing key column %s", key)
			}
			pk = pk + v
		}
		result[pk] = object
	}
	return result, nil
}
