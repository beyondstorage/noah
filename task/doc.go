package task

//go:generate go run ../internal/cmd/generator/tasks
//go:generate mockgen -package mock -destination ../pkg/mock/storager.go github.com/aos-dev/go-storage/v2/types Copier,DirLister,DirSegmentsLister,IndexSegmenter,Mover,PrefixLister,PrefixSegmentsLister,Reacher,Servicer,Statistician,Storager
//go:generate mockgen -package mock -destination ../pkg/mock/io.go io ReadCloser
