package workers

import (
	"fmt"
	"go.uber.org/zap"
)

type SelectionInfo struct {
	ReferenceSolution string `db:"REFERENCE_SOLUTION"`
	DefaultSchema     string `db:"DEFAULT_SCHEMA"`
	CheckOrder        string `db:"CHECK_ORDER"`
	CheckColumnNames  string `db:"CHECK_COLUMN_NAMES"`
	Schemas           []string
}

func (p *Pipeline) FetchSelectionInfo(s Submission) (Submission, error) {
	var info SelectionInfo

	err := p.mainDB.Get(&info, qFetchSelectionInfo, s.ID)
	if err != nil {
		return s, err
	}

	err = p.mainDB.Select(&info.Schemas, qFetchSelectionSchemas, s.ID)
	if err != nil {
		return s, err
	}

	info.ReferenceSolution = prepareSelectionSolution(info.ReferenceSolution)

	s.CheckerInfo = info

	return s, nil
}

func (p *Pipeline) CheckSelectionColumns(s Submission) (Submission, error) {
	selectionInfo := s.CheckerInfo.(SelectionInfo)

	solutionRows, err := p.selectionDBs[selectionInfo.DefaultSchema].Queryx(s.Solution)
	if err != nil {
		return s, err
	}

	referenceRows, err := p.selectionDBs[selectionInfo.DefaultSchema].Queryx(selectionInfo.ReferenceSolution)
	if err != nil {
		return s, err
	}

	solutionColumns, err := solutionRows.Columns()
	if err != nil {
		return s, err
	}

	referenceColumns, err := referenceRows.Columns()
	if err != nil {
		return s, err
	}

	if len(solutionColumns) != len(referenceColumns) {
		s.StatusID = IncorrectColumnCount
		return s, nil
	}

	for i := 0; i < len(solutionColumns); i++ {
		if solutionColumns[i] != referenceColumns[i] {
			s.StatusID = IncorrectColumnNames
			break
		}
	}

	if err := solutionRows.Close(); err != nil {
		return s, err
	}

	if err := referenceRows.Close(); err != nil {
		return s, err
	}

	return s, nil
}

func (p *Pipeline) CheckSelectionContent(s Submission) (Submission, error) {
	selectionInfo := s.CheckerInfo.(SelectionInfo)

	var q string
	if selectionInfo.CheckOrder == "Y" {
		q = qCheckOrderedContent
	} else {
		q = qCheckUnorderedContent
	}

	query := fmt.Sprintf(q, s.Solution, selectionInfo.ReferenceSolution)

	// TODO: debug
	// log.Println(query)

	rows, err := p.selectionDBs[selectionInfo.DefaultSchema].Query(query)
	if err != nil {
		s.PutCheckerError(err)
		return s, err
	}

	if rows == nil || !rows.Next() {
		s.StatusID = Accepted
		return s, rows.Close()
	}

	if selectionInfo.CheckOrder == "Y" {
		s.StatusID = IncorrectOrder
	} else {
		s.StatusID = IncorrectContent
	}

	return s, rows.Close()
}

func (p *Pipeline) CheckSelectionSolution(s Submission) (Submission, error) {
	var err error

	s.Solution = prepareSelectionSolution(s.Solution)

	if s.CheckerInfo.(SelectionInfo).CheckColumnNames == "Y" {
		s, err = p.CheckSelectionColumns(s)
		if s.IsChecked() || err != nil {
			return s, err
		}
	}

	return p.CheckSelectionContent(s)
}

func (p *Pipeline) SelectionChecker(unchecked <-chan Submission) <-chan Submission {
	check := func(s Submission) Submission {
		s, err := p.FetchSelectionInfo(s)
		if err != nil {
			s.PutCheckerError(err)
			p.logger.Error("failed to fetch selection checker info", zap.Error(err))
			return s
		}

		s, err = p.CheckSelectionSolution(s)
		if err != nil {
			s.PutCheckerError(err)
			p.logger.Error("failed to check selection solution", zap.Error(err))
		}

		return s
	}

	return p.Checker("selection", unchecked, check)
}
