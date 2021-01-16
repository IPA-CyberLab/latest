package parser

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/blang/semver/v4"

	"github.com/IPA-CyberLab/latest/pkg/query"
)

var reStripPrefix = regexp.MustCompile(`^[A-z_\-]*`)

func ParseVersion(s string) (semver.Version, error) {
	s = reStripPrefix.ReplaceAllString(s, "")

	ver, err := semver.ParseTolerant(s)
	if err != nil {
		return semver.Version{}, nil
	}

	return ver, nil
}

type queryIntermediate struct {
	SoftwareId  string
	VerRangeStr string
	Prerelease  bool
}

var reSoftwareIdAndRest = regexp.MustCompile(`^([^:@<>=]*)(.*)$`)
var reAtVersion = regexp.MustCompile(`^@v?(\d+)(\.(\d+))?(\.(\d+))?(.*)$`)
var reRangeVersion = regexp.MustCompile(`^([<>]=?[\d\.]+)(.*)$`)
var reFlag = regexp.MustCompile(`^:(\w+)(.*)$`)

func parseInternal(s string) (*queryIntermediate, error) {
	ms := reSoftwareIdAndRest.FindStringSubmatch(s)
	if len(ms) == 0 {
		log.Panicf("Unexpected to match nothing: %q", s)
	}

	softwareId, attrstr := ms[1], ms[2]

	qi := &queryIntermediate{
		SoftwareId: softwareId,
	}

	for attrstr != "" {
		ms := reAtVersion.FindStringSubmatch(attrstr)
		if len(ms) != 0 {
			major, minor, patch := ms[1], ms[3], ms[5]
			if patch != "" {
				qi.VerRangeStr = fmt.Sprintf("=%s.%s.%s %s", major, minor, patch, qi.VerRangeStr)
			} else if minor != "" {
				nminor, err := strconv.Atoi(minor)
				if err != nil {
					log.Panicf("Unexpected Atoi failure: %q", major)
				}
				qi.VerRangeStr = fmt.Sprintf(">=%[1]s.%[2]s.0 <%[1]s.%[3]d.0 %[4]s", major, minor, nminor+1, qi.VerRangeStr)
			} else {
				nmajor, err := strconv.Atoi(major)
				if err != nil {
					log.Panicf("Unexpected Atoi failure: %q", major)
				}
				qi.VerRangeStr = fmt.Sprintf(">=%[1]s.0.0 <%[2]d.0.0 %[3]s", major, nmajor+1, qi.VerRangeStr)
			}

			attrstr = ms[6]
			continue
		}

		ms = reRangeVersion.FindStringSubmatch(attrstr)
		if len(ms) != 0 {
			rangestr := ms[1]
			log.Printf("range: %q", rangestr)

			qi.VerRangeStr = fmt.Sprintf("%s %s", rangestr, qi.VerRangeStr)

			attrstr = ms[2]
			continue
		}

		ms = reFlag.FindStringSubmatch(attrstr)
		if len(ms) != 0 {
			flag := ms[1]
			log.Printf("flag: %q", flag)

			switch flag {
			case "prerelease":
				qi.Prerelease = true
			default:
				return nil, fmt.Errorf("Unknown flag attribute: %q", flag)
			}

			attrstr = ms[2]
			continue
		}

		return nil, fmt.Errorf("Failed to parse attrs %q", attrstr)
	}

	return qi, nil
}

func Parse(s string) (*query.Query, error) {
	qi, err := parseInternal(s)
	if err != nil {
		return nil, err
	}

	var vr semver.Range
	if qi.VerRangeStr != "" {
		vr, err = semver.ParseRange(qi.VerRangeStr)
		if err != nil {
			return nil, err
		}
	}

	q := &query.Query{
		SoftwareId: qi.SoftwareId,
		VerRange:   vr,
		Prerelease: qi.Prerelease,
	}
	return q, nil
}
