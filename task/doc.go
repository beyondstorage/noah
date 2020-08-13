package task

//go:generate go run ../internal/cmd/generator/tasks
//go:generate mockgen -package mock -destination ../pkg/mock/storager.go github.com/aos-dev/go-storage/v2 Storager,Servicer,Reacher,PrefixSegmentsLister,IndexSegmenter,DirLister,PrefixLister,Statistician
//go:generate mockgen -package mock -destination ../pkg/mock/io.go io ReadCloser
