package provider_test

import (
	"os"
	"testing"
)

func requireAcceptance(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run Sigma acceptance tests")
	}
	t.Skip("acceptance test requires dedicated Sigma test fixtures")
}

func TestAccMemberResource(t *testing.T)                      { requireAcceptance(t) }
func TestAccTeamResource(t *testing.T)                        { requireAcceptance(t) }
func TestAccTeamMemberResource(t *testing.T)                  { requireAcceptance(t) }
func TestAccTeamMembersResource(t *testing.T)                 { requireAcceptance(t) }
func TestAccAccountTypeResource(t *testing.T)                 { requireAcceptance(t) }
func TestAccUserAttributeResource(t *testing.T)               { requireAcceptance(t) }
func TestAccUserAttributeTeamAssignmentResource(t *testing.T) { requireAcceptance(t) }
func TestAccUserAttributeUserAssignmentResource(t *testing.T) { requireAcceptance(t) }
func TestAccWorkspaceResource(t *testing.T)                   { requireAcceptance(t) }
func TestAccWorkspaceGrantResource(t *testing.T)              { requireAcceptance(t) }
func TestAccFileResource(t *testing.T)                        { requireAcceptance(t) }
func TestAccGrantResource(t *testing.T)                       { requireAcceptance(t) }
func TestAccWorkbookGrantResource(t *testing.T)               { requireAcceptance(t) }
func TestAccReportGrantResource(t *testing.T)                 { requireAcceptance(t) }
func TestAccConnectionResource(t *testing.T)                  { requireAcceptance(t) }
func TestAccConnectionGrantResource(t *testing.T)             { requireAcceptance(t) }
func TestAccConnectionPathGrantResource(t *testing.T)         { requireAcceptance(t) }
func TestAccAPIConnectorResource(t *testing.T)                { requireAcceptance(t) }
func TestAccAPICredentialResource(t *testing.T)               { requireAcceptance(t) }
func TestAccTagResource(t *testing.T)                         { requireAcceptance(t) }
func TestAccWorkbookScheduleResource(t *testing.T)            { requireAcceptance(t) }
func TestAccReportScheduleResource(t *testing.T)              { requireAcceptance(t) }
func TestAccWorkbookEmbedResource(t *testing.T)               { requireAcceptance(t) }
func TestAccTranslationResource(t *testing.T)                 { requireAcceptance(t) }
