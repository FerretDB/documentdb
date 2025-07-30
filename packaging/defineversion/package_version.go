package main

import (
	"fmt"
	"os"
	"regexp"
)

// controlDefaultVer matches major, minor and "patch" from `default_version` field in control file,
// see pg_documentdb_core/documentdb_core.control.
var controlDefaultVer = regexp.MustCompile(`(?m)^default_version = '(?P<major>[0-9]+)\.(?P<minor>[0-9]+)-(?P<patch>[0-9]+)'$`)

// disallowedDebian matches disallowed characters of Debian `upstream_version` when used without `debian_revision`.
// See https://www.debian.org/doc/debian-policy/ch-controlfields.html#version.
var disallowedDebian = regexp.MustCompile(`[^A-Za-z0-9\.+~]`)

// disallowedRPM matches disallowed characters of pre-release string.
// See https://fedoraproject.org/wiki/PackagingDrafts/TildeVersioning#Basic_versioning_rules.
var disallowedRPM = regexp.MustCompile(`[^A-Za-z0-9\._+~]`)

// getControlDefaultVersion returns the default_version field from the control file
// in SemVer format (0.103-0 -> 0.103.0).
func getControlDefaultVersion(f string) (string, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return "", err
	}

	match := controlDefaultVer.FindSubmatch(b)
	if match == nil || len(match) != controlDefaultVer.NumSubexp()+1 {
		return "", fmt.Errorf("control file did not find default_version: %s", f)
	}

	major := match[controlDefaultVer.SubexpIndex("major")]
	minor := match[controlDefaultVer.SubexpIndex("minor")]
	patch := match[controlDefaultVer.SubexpIndex("patch")]

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}
