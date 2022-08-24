package util

import (
	"net/url"
	"path/filepath"
	"strings"

	"syslabit.com/git/syslabit/log"
)

func GetRemoteRepoBranches(uri, login, password string) ([]string, *log.RecordS) {

	out, err := ExecCommand("git", "ls-remote", "--heads", uri)
	if err == nil {
		// public repo
		return getBranches(out), nil
	}

	// private repo
	if login == "" || password == "" {
		return nil, log.Error("login and password are required")
	}

	// build url https://user:password@github.com/...
	tmp := strings.Split(uri, "://")
	uri = tmp[0] + "://" + login + ":" + url.QueryEscape(password) + "@" + strings.Join(tmp[1:], "")

	// check access with login and password
	out, err = ExecCommand("git", "ls-remote", "--heads", uri)
	if err != nil {
		log.Error("remote project git ls-remote fail", log.Vars{
			"exit":       err,
			"cmd-output": out,
			"url":        uri,
		})
		return nil, log.Error("login and password are invalid")
	}

	return getBranches(out), nil
}

func getBranches(cmdOut string) []string {
	branches := []string{}

	splitted := strings.Split(cmdOut, "\n")
	length := len(splitted)
	for i, line := range splitted {
		if i == length-1 {
			break
		}
		splitted := strings.Split(line, "refs/heads/")
		if len(splitted) > 1 {
			branches = append(branches, splitted[1])
		}
	}
	return branches
}

func GetCommitCount(gitPath, projectName, commit string) string {

	cmd := []string{
		"git",
		"--no-pager",
		"-C", filepath.Join(gitPath, projectName), // git repo dir
		"rev-list",
		"--count",
		commit,
	}

	out, err := ExecCommand(cmd...)

	if err != nil {
		log.Error(out+" "+err.Error(), log.Vars{
			"project":    projectName,
			"cmd":        strings.Join(cmd, " "),
			"cmd-output": out,
		})
		return ""
	}

	return strings.TrimSpace(out)
}
