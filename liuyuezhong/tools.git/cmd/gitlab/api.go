package main

import (
	"fmt"
	"github.com/popstk/inke/internal/utils"
	"github.com/xanzy/go-gitlab"
)

func ListGroupProjects(git *gitlab.Client, root interface{}, f func(project *gitlab.Project) error) error {
	lo := gitlab.ListOptions{
		Page:    1,
		PerPage: 50,
	}
	for {
		projects, rsp, err := git.Groups.ListGroupProjects(root, &gitlab.ListGroupProjectsOptions{
			ListOptions:      lo,
			IncludeSubgroups: gitlab.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("ListGroupProjects err: %v", err)
		}

		if verbose {
			utils.PrintWarn("ListGroupProjects gid[%v] rsp: %+v", root, rsp)
		}

		for _, p := range projects {
			if err := f(p); err != nil {
				return err
			}
		}

		if rsp.NextPage == 0 {
			break
		}
		lo.Page = rsp.NextPage
	}

	return nil
}
