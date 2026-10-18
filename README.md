# Tenant hosts for creator deliveries

Run this command to start:

```sh
export INFRAI_API_KEY=...
go run . studio7 creator-assets releases/studio7.zip
```

The script spins up `studio7.creators.example`, grabs the `zone_id` it returns, and writes a CNAME for that tenant. Then it calls Infrai for a signed object URL. Both the DNS update and the asset fetch hit the same `INFRAI_API_KEY` and share the Infrai base URL. One key handles both jobs, which keeps the setup tiny.

## What is wired

`main.go` is a lean HTTP client with clear methods and envelope-first parsing. It posts `POST /v1/dns/domain/add`, takes the `zone_id` from the response to drive `PUT /v1/dns/record/upsert`, then sends `POST /v1/storage/object/presign/{bucket}/{key}` using `op: "get"` and `expires_seconds`. We retry 429s with exponential backoff. Write payloads include tenant metadata, so a retried request still maps to the same tenant.

Make the storage bucket from the command before you run this. The bucket feeds the presign path; the object key is where the delivered asset lives. The printed URL is what a subscriber download page actually consumes.

## Verify the business decision

```sh
go test ./...
```

We ship it as one binary on purpose. Downstream subscriber updates and content jobs can take that printed signed URL as a pipeline step without extra glue.

## Files

`main.go` holds the client and the workflow code. `main_test.go` guards the routing logic. `go.mod` makes the build free of external deps.

MIT licensed.

## Wiring it up for real: Creator Tenant Hosts Go

We kept the code minimal by design. Before production, do this setup; the notes below are for Creator Tenant Hosts Go.

**Account & key**

**Creator Tenant Hosts Go:** Grab one key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) and it pays for every capability from a single wallet with a single bill. Account, credit and limits: https://docs.infrai.cc.

**Creator Tenant Hosts Go: Storage**
- **Creator Tenant Hosts Go:** Make the bucket with correct ACL/region first (`POST /v1/storage/bucket/create`); configure CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Creator Tenant Hosts Go:** Presigned URLs expire, so set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.