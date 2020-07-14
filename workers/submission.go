package workers

import (
	"fmt"
	"go.uber.org/zap"
)

const (
	Unknown int = iota
	PendingCheck
	Accepted
	ExecutionError
	RestrictionViolated
	IncorrectColumnCount
	IncorrectColumnNames
	IncorrectContent
	IncorrectOrder
)

type Submission struct {
	ID             int    `db:"ID"`
	StatusID       int    `db:"STATUS_ID"`
	Solution       string `db:"SOLUTION"`
	CheckerMessage string `db:"CHECKER_MESSAGE"`

	// used by various checkers
	CheckerInfo interface{}
}

func (s *Submission) IsChecked() bool {
	return s.StatusID != PendingCheck
}

func (s *Submission) PutCheckerError(err error) {
	s.StatusID = ExecutionError
	s.CheckerMessage = fmt.Sprintf("checker error: %s", err)

	// TODO: debug
	// log.Printf("%s %s", s.Solution, s.CheckerInfo.(SelectionInfo).ReferenceSolution)
}

func (p *Pipeline) SubmissionProvider() <-chan Submission {
	submissions := make(chan Submission)

	go func() {
		defer func() {
			p.logger.Info("stopped submission provider")
			close(submissions)
		}()
		for {
			select {
			case <-p.done:
				return
			case <-p.fetchTicker.C:
				rows, err := p.mainDB.Queryx(qFetchUncheckedSubmissions)
				if err != nil {
					p.logger.Error("failed to fetch unchecked submissions", zap.Error(err))
				}

				var s Submission
				for rows.Next() {
					if err := rows.StructScan(&s); err != nil {
						p.logger.Error("failed to scan submission into struct", zap.Error(err))
					} else {
						submissions <- s
					}
				}
			}
		}
	}()

	p.logger.Info("started submission provider")

	return submissions
}
