// The identity contract's public surface. gen/ is generated and never edited;
// this file is what other packages import.
export {
	PublicService,
	type RedeemEnrolmentLinkRequest,
	type RedeemEnrolmentLinkResponse,
	type BeginEnrolmentRequest,
	type BeginEnrolmentResponse,
	type FinishEnrolmentRequest,
	type FinishEnrolmentResponse,
	type BeginSignInRequest,
	type BeginSignInResponse,
	type FinishSignInRequest,
	type FinishSignInResponse,
	type SignOutRequest,
	type SignOutResponse,
	type GetCurrentMemberRequest,
	type GetCurrentMemberResponse
} from '../gen/ts/wimm/identity/v1/public_pb.js';

export {
	OperatorService,
	type RegisterMemberRequest,
	type RegisterMemberResponse,
	type IssueEnrolmentLinkRequest,
	type IssueEnrolmentLinkResponse
} from '../gen/ts/wimm/identity/v1/operator_pb.js';

export {
	EnrolmentFailure,
	EnrolmentLinkRejection,
	type EmailAddress,
	type EnrolmentLink,
	type Member
} from '../gen/ts/wimm/identity/v1/identity_pb.js';
