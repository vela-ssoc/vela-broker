package agtsvc

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/gridfs"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
)

type ThirdService interface {
	Open(ctx context.Context, name string) (*model.Third, gridfs.File, error)
}

func Third(qry *query.Query, gfs gridfs.FS) ThirdService {
	return &thirdService{
		qry: qry,
		gfs: gfs,
	}
}

type thirdService struct {
	qry *query.Query
	gfs gridfs.FS
}

func (biz *thirdService) Open(ctx context.Context, name string) (*model.Third, gridfs.File, error) {
	tbl := biz.qry.Third
	th, err := tbl.WithContext(ctx).Where(tbl.Name.Eq(name)).First()
	if err != nil {
		return nil, nil, err
	}

	file, err := biz.gfs.OpenID(th.FileID)
	if err != nil {
		return nil, nil, err
	}

	return th, file, nil
}
