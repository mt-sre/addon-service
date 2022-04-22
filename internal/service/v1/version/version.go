package version

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	v1 "github.com/mt-sre/addon-service/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// https://github.com/golang/go/issues/37369
const (
	empty = "was not build properly"
)

// Values are provided by compile time -ldflags.
var (
	Version   = empty
	Branch    = empty
	Commit    = empty
	BuildDate = empty
)

type VersionService struct {
}

func NewVersionService() *VersionService {
	return &VersionService{}
}

// Get returns the build-in version and platform information
func (svc *VersionService) Get(_ context.Context, _ *emptypb.Empty) (*v1.Version, error) {
	v := &v1.Version{
		Version:   Version,
		Branch:    Branch,
		Commit:    Commit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if BuildDate != empty {
		i, err := strconv.ParseInt(BuildDate, 10, 64)
		if err != nil {
			return &v1.Version{}, fmt.Errorf("error parsing build time: %w", err)
		}
		v.BuildTime = timestamppb.New(time.Unix(i, 0).UTC())
	}
	return v, nil
}
