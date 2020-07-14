package workers

import (
	"fmt"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

type Restriction struct {
	Type       string `db:"RESTRICTION_TYPE"`
	Definition string `db:"DEFINITION"`
}

type RestrictionInfo []Restriction

func (p *Pipeline) FetchRestrictions(s Submission) (Submission, error) {
	var restrictions RestrictionInfo

	err := p.mainDB.Select(&restrictions, qFetchSubmissionRestrictions, s.ID)
	if err != nil {
		return s, err
	}

	s.CheckerInfo = restrictions

	return s, nil
}

func (p *Pipeline) CheckRestrictions(s Submission) (Submission, error) {
	switch t := s.CheckerInfo.(type) {
	case RestrictionInfo:
		for _, r := range s.CheckerInfo.(RestrictionInfo) {
			var err error
			violated := false

			switch r.Type {
			case "KEYWORD":
				violated = strings.Contains(s.Solution, r.Definition)
			case "REGEXP":
				violated, err = regexp.MatchString(r.Definition, s.Solution)
			default:
				return s, fmt.Errorf("failed to check restriction of type \"%s\"", r.Type)
			}

			if violated {
				s.StatusID = RestrictionViolated
				return s, err
			}
		}
	default:
		return s, fmt.Errorf("invalid checker info type: %T", t)
	}

	return s, nil
}

func (p *Pipeline) RestrictionChecker(unchecked <-chan Submission) <-chan Submission {
	check := func(s Submission) Submission {
		s, err := p.FetchRestrictions(s)
		if err != nil {
			p.logger.Error("failed to fetch restrictions", zap.Error(err))
			return s
		}

		s, err = p.CheckRestrictions(s)
		if err != nil {
			p.logger.Error("failed to check restrictions", zap.Error(err))
		}

		return s
	}

	return p.Checker("restriction", unchecked, check)
}
