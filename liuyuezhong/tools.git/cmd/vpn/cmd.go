package main

import (
	"context"
	"fmt"
	"github.com/kardianos/service"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func SysLog(s service.Service) service.Logger {
	errs := make(chan error, 5)
	logger, err := s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	return logger
}

func (p *Program) Start(s service.Service) error {
	ctx, cancel := context.WithCancel(context.TODO())
	p.cancel = cancel

	if !service.Interactive() {
		log.SetOutput(p.ProgramLog)
	}

	go p.monitor(ctx)
	return nil
}

func (p *Program) monitor(ctx context.Context) {
	const minDuration = 15 * time.Second
	duration := minDuration

	for {
		if err := p.run(ctx); err != nil {
			log.Printf("[host] run err: %v", err)

			duration *= 2
			if duration > time.Hour {
				duration = time.Hour
			}
		} else {
			duration = minDuration
		}

		timer := time.NewTimer(duration)
		log.Printf("timer %v", duration)

		select {
		case <-ctx.Done():
			log.Printf("[host] exit")
			timer.Stop()
			return
		case <-timer.C:
			break
		}
	}
}

func (p *Program) run(ctx context.Context) error {
	log.Printf("run ...")

	conf, err := ParseConf(p.file)
	if err != nil {
		return fmt.Errorf("ParseConf err: %v", err)
	}

	var builder strings.Builder
	if conf.ServerCert != "" {
		builder.WriteString("--servercert ")
		builder.WriteString(conf.ServerCert)
		builder.WriteByte(' ')
	}

	builder.WriteString("-u ")
	builder.WriteString(conf.UserName)
	builder.WriteByte(' ')

	for field, entry := range conf.FormEntry {
		builder.WriteString("--form-entry main:")
		builder.WriteString(field)
		builder.WriteByte('=')
		builder.WriteString(GetSecret(entry.Type, entry.Secret))
		builder.WriteByte(' ')
	}

	if verbose {
		builder.WriteString("--dump ")
	}

	builder.WriteString(conf.Server)

	fmt.Printf("cmdline: [%s]\n", builder.String())
	c := exec.CommandContext(ctx, conf.Program,
		strings.Split(builder.String(), " ")...)

	if service.Interactive() {
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
	} else {
		c.Stderr = p.ProgramLog
		c.Stdout = p.ProgramLog
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("start err: %v", err)
	}

	if err := c.Wait(); err != nil {
		return fmt.Errorf("wait err: %v", err)
	}

	return nil
}

// Stop should not block. Return with a few seconds.
func (p *Program) Stop(s service.Service) error {
	p.cancel()
	return nil
}

type Program struct {
	file       string
	ProgramLog io.Writer
	cancel     context.CancelFunc
}

func NewProgram(file string) *Program {
	path, err := os.Executable()
	if err != nil {
		log.Fatalf("os.Executable err: %v", err)
	}

	path += ".log"
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("os.Open err: %v", err)
	}

	conf := filepath.Join(filepath.Dir(path), file)

	return &Program{
		file:       conf,
		ProgramLog: NewLockedWriter(f),
	}
}
