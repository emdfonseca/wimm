// Command wimmctl is the operator's command line for this wimm instance.
//
// It reaches the service over Connect rather than the database, so there is
// exactly one place an enrolment link can be minted (ADR 0016).
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"

	identityv1 "github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1"
	"github.com/xuuid/wimm/packages/contracts/gen/go/wimm/identity/v1/identityv1connect"

	"github.com/xuuid/wimm/apps/wimm/internal/config"
	"github.com/xuuid/wimm/apps/wimm/internal/rpc"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

const usage = `wimmctl — operator commands for a wimm instance

  wimmctl register <email> <first name> <last name>
        Register a household member and print a link to send them.

  wimmctl link <email>
        Issue a further enrolment link to someone already registered.
        Any link they are still holding stops working; their passkeys do not.

Environment:
  WIMM_OPERATOR_ADDR        the operator listener (default %s)
  WIMM_OPERATOR_CREDENTIAL  the credential the instance is configured with
`

func run(args []string, stdout, stderr *os.File) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, usage, config.DefaultOperatorAddr)
		return 2
	}

	addr := os.Getenv("WIMM_OPERATOR_ADDR")
	if addr == "" {
		addr = config.DefaultOperatorAddr
	}
	credential := os.Getenv("WIMM_OPERATOR_CREDENTIAL")
	if credential == "" {
		fmt.Fprintln(stderr, "WIMM_OPERATOR_CREDENTIAL is not set: it must match what the instance is configured with")
		return 2
	}

	client := identityv1connect.NewOperatorServiceClient(http.DefaultClient, "http://"+addr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch args[0] {
	case "register":
		return register(ctx, client, credential, args[1:], stdout, stderr)
	case "link":
		return issueLink(ctx, client, credential, args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprintf(stdout, usage, config.DefaultOperatorAddr)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		fmt.Fprintf(stderr, usage, config.DefaultOperatorAddr)
		return 2
	}
}

func register(ctx context.Context, client identityv1connect.OperatorServiceClient, credential string, args []string, stdout, stderr *os.File) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: wimmctl register <email> <first name> <last name>")
		return 2
	}

	req := connect.NewRequest(&identityv1.RegisterMemberRequest{
		Email:     &identityv1.EmailAddress{Value: args[0]},
		FirstName: args[1],
		LastName:  args[2],
	})
	req.Header().Set(rpc.OperatorCredentialHeader, "Bearer "+credential)

	res, err := client.RegisterMember(ctx, req)
	if err != nil {
		return reportFailure(err, stderr)
	}

	m := res.Msg.GetMember()
	fmt.Fprintf(stdout, "Registered %s %s <%s>.\n\n", m.GetFirstName(), m.GetLastName(), m.GetEmail())
	printLink(stdout, res.Msg.GetEnrolmentLink())
	return 0
}

func issueLink(ctx context.Context, client identityv1connect.OperatorServiceClient, credential string, args []string, stdout, stderr *os.File) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: wimmctl link <email>")
		return 2
	}

	req := connect.NewRequest(&identityv1.IssueEnrolmentLinkRequest{
		Email: &identityv1.EmailAddress{Value: args[0]},
	})
	req.Header().Set(rpc.OperatorCredentialHeader, "Bearer "+credential)

	res, err := client.IssueEnrolmentLink(ctx, req)
	if err != nil {
		return reportFailure(err, stderr)
	}

	m := res.Msg.GetMember()
	fmt.Fprintf(stdout, "New link for %s %s <%s>. Any earlier link has stopped working; their passkeys have not.\n\n",
		m.GetFirstName(), m.GetLastName(), m.GetEmail())
	printLink(stdout, res.Msg.GetEnrolmentLink())
	return 0
}

// printLink shows the link and when it stops working, so the operator can say
// so when they send it. This is the only time the value is ever shown.
func printLink(stdout *os.File, link *identityv1.EnrolmentLink) {
	expires := link.GetExpiresAt().AsTime().Local()
	fmt.Fprintf(stdout, "  %s\n\n", link.GetUrl())
	fmt.Fprintf(stdout, "It stops working at %s.\n", expires.Format("15:04 on 2 January 2006"))
	fmt.Fprintln(stdout, "Send it over a channel you trust: whoever opens it becomes this member.")
}

func reportFailure(err error, stderr *os.File) int {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		fmt.Fprintf(stderr, "%s\n", connectErr.Message())
		if connectErr.Code() == connect.CodeUnauthenticated {
			fmt.Fprintln(stderr, "Check WIMM_OPERATOR_CREDENTIAL against the instance's configuration.")
		}
		if connectErr.Code() == connect.CodeUnavailable {
			fmt.Fprintln(stderr, "Check that wimmd is running and that WIMM_OPERATOR_ADDR points at its operator listener.")
		}
		return 1
	}
	fmt.Fprintf(stderr, "%v\n", err)
	return 1
}
