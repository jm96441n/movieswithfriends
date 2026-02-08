package http

// Party member handlers live here.
//
// Note: The AcceptInviteHandler (POST /party_members) is currently in invitations.go
// because accepting an invitation is part of the invitation flow and creates a party member.
// If future party member operations are added (e.g., remove member, leave party),
// they should be implemented here and reference the AcceptInviteHandler if needed.
