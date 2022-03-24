package source

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
)

type ExcelScanner struct {
	Path, Sheet string
	rows        *excelize.Rows
}

func (e *ExcelScanner) Open() error {
	f, err := excelize.OpenFile(e.Path)
	if err != nil {
		return fmt.Errorf("excelize.OpenFile Path %s err: %w", e.Path, err)
	}

	e.rows, err = f.Rows(e.Sheet)
	if err != nil {
		return fmt.Errorf("rows err: %w", err)
	}

	return nil
}

func (e *ExcelScanner) Next() bool {
	return e.rows.Next()
}

func (e *ExcelScanner) Record() []string {
	return e.rows.Columns()
}

func (e *ExcelScanner) Close() error {
	return nil
}

func (e *ExcelScanner) Err() error {
	return e.rows.Error()
}

func FromExcel(path, sheet string, fun Action, options ...ScanOptions) error {
	scanner := &ExcelScanner{
		Path:  path,
		Sheet: sheet,
	}
	return FromScanner(scanner, fun, options...)
}
