package source

type ScanOption struct {
	Batch     int // 每次批量处理
	SkipLines int // 忽略前n行
}

type ScanOptions func(*ScanOption)

func WithBatch(batch int) ScanOptions {
	return func(option *ScanOption) {
		option.Batch = batch
	}
}

func WithSkipLines(skipLines int) ScanOptions {
	return func(option *ScanOption) {
		option.SkipLines = skipLines
	}
}

type Action func(record [][]string)
