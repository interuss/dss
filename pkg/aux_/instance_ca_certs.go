package aux

import (
	"context"
	"fmt"
	"os"
	"strings"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/dss/pkg/datastore/flags"
)

func (a *Server) GetInstanceCAs(ctx context.Context, req *restapi.GetInstanceCAsRequest) restapi.GetInstanceCAsResponseSet {

	connectParameters := flags.ConnectParameters()

	CAFile := connectParameters.GetInstanceCAFile()

	if CAFile == "" {
		return restapi.GetInstanceCAsResponseSet{Response200: &restapi.CAsResponse{}}
	}

	data, err := os.ReadFile(CAFile)

	if err != nil {
		msg := fmt.Sprintf("Unable to read CA certificate file for this DSS instance, did try to read '%s', got: %s", CAFile, err)
		return restapi.GetInstanceCAsResponseSet{Response501: &restapi.ErrorResponse{Message: &msg}}
	}

	var CAs []string

	for _, CA := range strings.Split(string(data), START_OF_CERTIFICATE) {
		CA = strings.Trim(CA, "\r\n")

		if CA != "" {
			// Re-add the start mark that was removed by Split()
			CA = START_OF_CERTIFICATE + "\n" + CA
			CAs = append(CAs, CA)
		}
	}

	return restapi.GetInstanceCAsResponseSet{Response200: &restapi.CAsResponse{Cas: &CAs}}
}
