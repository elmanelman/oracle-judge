package workers

import (
	"github.com/elmanelman/oracle-judge/config"
	"github.com/elmanelman/oracle-judge/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type Pipeline struct {
	logger       *zap.Logger
	mainDB       *sqlx.DB
	selectionDBs map[string]*sqlx.DB

	fetchTicker *time.Ticker
	done        chan struct{}
}

func NewPipeline(cfg config.Config) (*Pipeline, error) {
	var p Pipeline
	var err error

	if p.logger, err = cfg.LoggerConfig.Build(); err != nil {
		return nil, err
	}
	if p.mainDB, err = database.ConnectWithConfig(cfg.MainDBConfig); err != nil {
		return nil, err
	}

	p.selectionDBs = make(map[string]*sqlx.DB, len(cfg.SelectionDBsConfigs))
	for name, c := range cfg.SelectionDBsConfigs {
		selectionDB, err := database.ConnectWithConfig(c)
		if err != nil {
			return nil, err
		}
		p.selectionDBs[name] = selectionDB
	}

	p.fetchTicker = time.NewTicker(time.Duration(cfg.CheckingConfig.FetchPeriod) * time.Millisecond)
	p.done = make(chan struct{})

	return &p, nil
}

func (p *Pipeline) Start() error {
	p.logger.Info("pipeline started")

	checked := p.SelectionChecker(p.RestrictionChecker(p.SubmissionProvider()))

	for submission := range checked {
		if err := p.UpdateSubmission(submission); err != nil {
			p.logger.Error("failed to update submission", zap.Error(err))
		}
	}

	return nil
}

func (p *Pipeline) Stop() {
	p.logger.Info("pipeline stopped")

	close(p.done)
}

func (p *Pipeline) UpdateSubmission(s Submission) error {
	_, err := p.mainDB.Exec(qUpdateSubmission, s.StatusID, s.CheckerMessage, s.ID)

	return err
}

func (p *Pipeline) Checker(kind string, unchecked <-chan Submission, check func(s Submission) Submission) <-chan Submission {
	checked := make(chan Submission)

	go func() {
		defer func() {
			p.logger.Info("stopped checker", zap.String("kind", kind))
			close(checked)
		}()
		for s := range unchecked {
			if !s.IsChecked() {
				s = check(s)
			}

			checked <- s
		}
	}()

	p.logger.Info("started checker", zap.String("kind", kind))

	return checked
}
