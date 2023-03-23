
---
authors: Michael Myers (michael.myers@goteleport.com)
state: draft
---

# RFD 113 - Web Moderated Session File Transfer

## Required approvers

Engineering: @zmb3 @xacrimon @jakule

## What

Approval process to allow file transfers during a web moderated session

## Why

Users have no ability to transfer files if their role requires a moderated session for the specified server, even if they are in an active and "valid" session currently. Originally, we would allow file transfers regardless of moderated sessions which, while good for UX, goes against the purpose of moderated sessions in the first place. We have disabled the functionality but would like to reintroduce it in a secure, moderated, and auditable way. 

## Scope

This RFD will cover how to implement for web SCP specifically, but may be expanded to SFTP for tsh, although my research has been mostly for web scp. If at anytime there is ambiguity in this RFD for "which type of file transfers does this refer to", default to web scp only. 

## The current issue

When a user initiates a file transfer request (HTTP request and `tcp scp`), the system establishes a new non-interactive session to run the "exec" command. When we attempt to [open the session](https://github.com/gravitational/teleport/blob/64b10f1ccbeeab63c4ce91f5f11fdd74b42448d8/lib/srv/sess.go#L277) we have a check named `checkIfStart` which will pull any policies that require moderation from the current identity, and return if they've been fulfilled or not. This will always fail as this is a "new" session, created just a few lines before the check itself. Therefore, we need a way to let the server know that this request has been moderator approved.

### Details

#### The FileTransferRequest 

The general idea of the solution is to provide a way for users to "request-a-request" that can be approved by moderators. An example of a file transfer specific struct could look like

```go
type FileTransferRequest {
    id string // UUID
    user string
    sessionID string
    filePath string
    approvers []*party
    sink bool // upload or download
}
```

We can add any relevant information otherwise needed. File-size was considered, but imagine a user trying to download a log file that could have changed size by the time the request flow has completed. 

Because this struct is going to allow an exec command to go through, we could abstract it further to _any_ SSHRequest being passed instead with a struct like below

```go
type ModeratedSSHRequest {
    id string // UUID
    user string
    sessionID string
    approvers []*party
    sshRequest *ssh.Request
}
```

`ssh.Request` contains a payload that can be matched against the incoming payload to verify it's the exact request that was approved.

This struct would be stored in the active session:

```go
type session struct {
    // ...
    approvedRequests map[string]*ApprovedSSHRequest
}
```

The benefit of storing it on the session, and not in an access-request-like object, is that once the session is gone, so is the approved request. Keeping these approvals as ephemeral as possible is ideal. 

#### the fileTransferChannel 

"How do we we send a FileTransferRequest anyway?". 
We will add a new channel and event in `web/terminal.go` named `fileTransferC` Similar to `resizeC`, we can send an event+payload envelope with the current websocket implementation. The client will implement a new  `MessageTypeEnum` that can be matched in the backend. "t" for transfer!

```diff
export const MessageTypeEnum = {
  RAW: 'r',
+ FILE_TRANSFER_REQUEST: 't',
  AUDIT: 'a',
  SESSION_DATA: 's',
  SESSION_END: 'c',
  RESIZE: 'w',
  WEBAUTHN_CHALLENGE: 'n',
};
```

`terminal.go` can then listen specifically for `FILE_TRANSFER_REQUEST` and handle accordingly.
1. Create a new `FileTransferRequest` struct
3. Broadcast a message to alert users (`"Teleport > michael has requested to download ~/myfile.txt"`)
4. Similar to how we send a `"window-change"` event, we would send a new event 
	1. `_, err := s.ch.SendRequest("file-transfer-request", false, Marshal(&req))`
5. Handle the request on the sshserver:
	1. send an audit event similar to the broadcast above
	2. store the request in the session
	3. Notify all members of the party (besides originator) of the request via the session connection. 

The reason we notify _all_ party members (originator excluded) is that a session might not necessarily require a moderator but just a second-party such as a peer. This would allow those types of sessions to still be audited and go through the file request approval process.

After the steps above we are in the `PENDING` state. This is not a blocking state and essentially just creates the request, and updates the relevant UI (more on the UI below). 

#### The Approval process

Similar to sending the `FileTransferRequest` event, another event will be listened to, the `FileTranferRequestResponse` (working name). We could separate the event into two with Approve/Deny but a simple switch case on a field should suffice. 

OnDeny:
	1. Remove request from the session struct. No need to keep it around in a "Denied" state, just delete.
	2. Emit audit event "XUser has denied ~/myfile.txt download request from michael".
	3. Broadcast message same as audit event above so the rest of the session can see
	4. Notify all members of the denial so the UI can be removed.

OnApprove:
	1. Add approver to the `approvers` slice
	2. Broadcast/audit event
	3. We can then use a policy checker to see if the approvers fulfill any moderation policy on the original requester. We can treat this check the same as the `checkIfStart` conditional for opening a session. If this comes back true, we notify the original requester with an event containing the ID of the `FileTransferRequest`

Once the client receives this final "approved" message, we can automatically send a "normal" SCP request (over HTTP) with two new optional params, `sessionID` and `requestID`. 

#### Updated file transfer api handler

Not much will change here. The only difference is adding the `requestID` and `sessionID` as env vars in the `serverContext`. This will allow us to extend our `SessionRegistry.OpenExecSession` method with a new check like `isApprovedRequest`

```go
func (s *SessionRegistry) isApprovedRequest(ctx context.Context, scx *ServerContext) (bool error) {
	// fetch session from registry with sessionID

	// find request in the session by requestID

	// validate incoming request (user/file/direction/etc) against the stored request

	// if all is good, TRUE
}
```

Then our `OpenExecSession` can be updated:

```go
func (s *SessionRegistry) OpenExecSession(ctx context.Context, channel ssh.Channel, scx *ServerContext) error {
	// ...
	if !canStart && !isApprovedRequest {
		return errCannotStartUnattendedSession
	}

	// ...
}
```

#### UX/UI

The request-a-request flow would be: the request -> the approval -> the file transfer

The `ApprovedFileTransferRequest` will be stored in-memory and not need to be persisted as it's own entity anywhere else.

We can extend our current File Transfer UI for almost every aspect of moderated file transfers with just a few additions.

A user will follow the normal file transfer dialog and, if in a moderated session, show a "waiting for approval state"

![new file transfer dialogs](https://user-images.githubusercontent.com/5201977/227379373-f7afb730-630f-4577-a12a-976279cc7cda.png)

If a request is approved, the user view will continue to look/function the same way as a regular request currently. 

#### Per-session-MFA

Once a file has been approved, it will go through the same process as file transfers now, so the "requester" gets all the MFA functionality for free and nothing will need to be done on that side.

We will have to add an MFA tap for approve/deny if MFA is required, but our sessions already have precence checks that require MFA so we can just reuse what already exists here.


### Security Considerations
Originally, an idea was thrown around to create a "signed" url with all the `FileTransferRequest` information encoded in it but this seemed unnecessary. Because the current flow stores the original request in the session, we aren't giving them full access to open any exec session, just one that matches the exact request that was asked for. Also, having the request stored in the session means that once the session is gone, there isn't a way to "re-request" it. 

I didn't speak above removing the request once it's been completed but that is a possibility. We can either
1. Remove the request once it's been approved and executed
2. Let the request live until the session is gone. This is less secure as it allows the user to download the file as many times as they want but would improve UX to not have to constantly ask for approval for the same file. I prefer option 1 and just take the UX hit in favor of security.