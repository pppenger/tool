package source

import "fmt"

func GetScanOptions(options ...ScanOptions) *ScanOption {
	o := &ScanOption{
		Batch:     100,
		SkipLines: 0,
	}
	for _, option := range options {
		option(o)
	}

	return o
}

type Scanner interface {
	Open() error
	Next() bool
	Record() []string
	Err() error
	Close() error
}

func FromScanner(scanner Scanner, fun Action, options ...ScanOptions) error {
	opt := GetScanOptions(options...)

	if err := scanner.Open(); err != nil {
		return fmt.Errorf("scanner open err: %w", err)
	}

	buff := make([][]string, 0, opt.Batch)
	for i := 0; scanner.Next(); i++ {
		record := scanner.Record()

		if i < opt.SkipLines {
			continue
		}

		buff = append(buff, record)
		if len(buff) >= opt.Batch {
			fmt.Println("process ", i+1)
			fun(buff)
			buff = buff[:0]
		}
	}

	if len(buff) > 0 {
		fun(buff)
	}

	return nil
}
