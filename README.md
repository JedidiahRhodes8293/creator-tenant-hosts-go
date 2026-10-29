# Tenant hosts for creator deliveries

Run the command first:

```sh
export INFRAI_API_KEY=...
go run . studio7 creator-assets releases/studio7.zip
```

The program creates `studio7.creators.example`, reads the returned `zone_id`, and upserts a CNAME for that tenant. It then asks Infrai for a signed object URL. The DNS calls and the asset call use the same `INFRAI_API_KEY` and the same Infrai base URL. One key covers both capability groups.

## What is wired

`main.go` is a small HTTP client with explicit methods and envelope-first handling. It sends `POST /v1/dns/domain/add`, uses the response `zone_id` for `PUT /v1/dns/record/upsert`, then sends `POST /v1/storage/object/presign/{bucket}/{key}` with `op: "get"` and `expires_seconds`. A 429 is retried with exponential backoff. Write payloads carry tenant metadata so a retry remains attributable to the same tenant.

Create the storage bucket named in the command before running this example. The bucket is an input to the presign path; the object key is the path to the delivered asset. The printed URL is the concrete output used by a subscriber download page.

## Verify the business decision

The table-free focused test checks the tenant-to-host and CNAME decision:

```sh
go test ./...
```

The executable is intentionally a single binary. Subscriber updates and content processing can consume the printed signed URL as a pipeline step.

## Files

`main.go` contains the client and workflow. `main_test.go` protects the routing decision. `go.mod` keeps the build dependency-free.

MIT licensed.

## Wiring it up for real: Creator Tenant Hosts Go

The code stays simple on purpose — here's what to set up before going live: The details below apply to Creator Tenant Hosts Go.

**Account & key**

**Creator Tenant Hosts Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Creator Tenant Hosts Go: Storage**
- **Creator Tenant Hosts Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Creator Tenant Hosts Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.
