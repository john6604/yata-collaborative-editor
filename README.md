# YATA CRDT Collaborative Text Editor

A real-time collaborative text editor where each client runs its own YATA algorithm instance, guaranteeing data convergence without central coordination. 
The relay server acts purely as a message forwarder that does not process, integrate or delete data. 

---

## Visual Demo

// GIF

---

## Architecture

Frontend is developed in `React + Vite` running the YATA engine directly in the browser. The relay server is developed in `Go` using websockets for bidirectional communication with the client, `BoltDB` is implemented as persistence for saving document's snapshots and `SQLite` used to store a document catalog.  

// Image

YATA runs directly on the browser which simplifies the deployment and eliminates the need for a `Go` process per client. The trade-offs are that local persistence is lost and CRDT may not survive if the browser is closed. 
The relay does not execute convergence logic because it would turn the application into a server dependent system, which is specifically what CRDT does not need because the data convergence comes from its mathematical properties.

---

## How it works?

// Image 1
Normal edition starts with the user applying an operation. Once the operation is fully applied, it is sent through websocket to the other connected clients reaching the same final state. 

// Image 2
When a client loses connection, the client can still edit the document. In parallel, other clients can send and receive operations. When the client finally reconnects, a state vector and a delete set are sent to the connected clients. The YATA algorithm computes what operations the recently reconnected client needs and send them. The client applies its pending operations and the mathematical property of CRDT is in charge of reaching the same state for every client. 

// Image 3
When a client connects for first time to a document, the other connected clients prepare a snapshot with the most recent information. The snapshot contains operations log and metadata, which allows the client to reconstruct the document applying the operations locally. . If no client is connected, the relay server has a stored snapshot ready to send to the new client. From that point, the new client can edit and send operations through websockets.

---

## Stack

| Technology | Use | Reason |
|------------|-----|-------------|
| Go | Backend | Ideal for managing concurrency in an application with multiple websocket connections |
| React +  Vite | Frontend | Fast development environment with reusable components |
| SQLite | Persistence | Embedded relational database that can run within an application without external servers | 
| BoltDB | Persistence | Embedded storage that does not require external database managers. Ideal for serialization |
| WebSocket | Communication | Native compatibility with browser. Permits bidirectional messages and persistent channels |
| Docker | Deployment | Simplifies deployment with a portable way to run applications in several computers |

---

## How to run

### Prerequisites

In order to start the application your computer will need to have `Docker` and `Git` installed. If any of those is not installed you can install them in the following links:
- [Docker](https://docs.docker.com/get-docker/)
- [Git](https://git-scm.com/downloads)

### Running the application

1. Clone the repository

`git clone https://github.com/john6604/yata-collaborative-editor.git`

2. Go to the root directory

`cd yata-collaborative-editor`

3. Build and run Docker

`docker compose up --build`

4. Open the URL in your browser

`http://localhost:8080`

If your port `8080` is occupied the application will fail to start. Update the port mapping in `docker-compose.yml`.

---

## Communication Protocol

// Protocolo de comunicacion diseñado para el proyecto
In order to allow communication between clients I have designed a specific message protocol that covers all operations within the application.

### Connection Envelope
 
Sent by the client to join a document room.
 
```json
{
  "version": 1,
  "type": "join",
  "payload": {
    "room": "document-id",
    "client_id": "client-uuid",
    "name": "username"
  }
}
```
 
### Operations Envelope
 
All messages after joining use this structure. The `type` field identifies the message kind, and `payload.op.type` identifies the specific operation.
 
#### Insert
 
Sent by a client when a character is inserted. The relay broadcasts it to all other clients in the room.
 
```json
{
  "version": 1,
  "type": "update",
  "payload": {
    "op": {
      "type": "insert",
      "new_id": { "client_id": "uuid", "clock": 0 },
      "origin_id": { "client_id": "START", "clock": -1 },
      "right_id": { "client_id": "END", "clock": -2 },
      "character": "H"
    }
  }
}
```
 
#### Delete
 
Sent by a client when a character is deleted.
 
```json
{
  "version": 1,
  "type": "update",
  "payload": {
    "op": {
      "type": "delete",
      "target_id": { "client_id": "uuid", "clock": 0 }
    }
  }
}
```
 
#### Sync Step 1
 
Sent by a client immediately after joining. Contains the client's current state vector and delete set so peers can compute what operations are missing.
 
```json
{
  "version": 1,
  "type": "update",
  "payload": {
    "op": {
      "type": "sync_step1",
      "vector_state": { "client-uuid-1": 5, "client-uuid-2": 3 },
      "delete_set": { "client-uuid-1": [2, 4] }
    }
  }
}
```
 
#### Sync Step 2
 
Sent by a peer (or the relay if no peers are available) in response to sync_step1. Contains the delta of missing operations. Fields inside the delta use PascalCase because they are serialized from Go structs, and `Content` is a Unicode code point (integer).
 
```json
{
  "version": 1,
  "type": "update",
  "payload": {
    "op": {
      "type": "sync_step2",
      "delta": {
        "inserts": [
          {
            "NewID": { "ClientID": "uuid", "Clock": 0 },
            "OriginID": { "ClientID": "START", "Clock": -1 },
            "RightID": { "ClientID": "END", "Clock": -2 },
            "Content": 72
          }
        ],
        "deletes": [
          { "TargetID": { "ClientID": "uuid", "Clock": 2 } }
        ]
      }
    }
  }
}
```
 
#### Snapshot
 
Sent by a client after operations are applied. The relay stores it as an opaque cache per document room to serve new clients when no peers are available.
 
```json
{
  "version": 1,
  "type": "update",
  "payload": {
    "op": {
      "type": "snapshot",
      "delta": {
        "inserts": [],
        "deletes": []
      }
    }
  }
}
```
 
#### Presence
 
Broadcast by the relay to all clients in a room when a user joins or leaves. Updates the connected users list.
 
```json
{
  "version": 1,
  "type": "update",
  "payload": {
    "op": {
      "type": "presence",
      "users": [
        { "id": "client-uuid-1", "name": "Alice" },
        { "id": "client-uuid-2", "name": "Bob" }
      ]
    }
  }
}
```
 
### Error Messages
 
Sent by the relay when a message cannot be processed.
 
```json
{
  "version": 1,
  "type": "error",
  "payload": {
    "code": "error_code",
    "message": "Human-readable description."
  }
}
```
 
| Error Code | Description |
|------------|-------------|
| `malformed_json` | The message could not be parsed as JSON |
| `unsupported_version` | The protocol version is not supported |
| `unknown_message_type` | The message type is not recognized |
| `missing_field` | A required field is missing from the message |
| `invalid_payload` | The payload structure is invalid |
| `not_joined` | The client has not joined a room yet |
| `unauthorized` | The client could not be identified |
| `payload_too_large` | The message exceeds the maximum allowed size |
| `internal_error` | An unexpected server error occurred |

---

## Design Decisions



---

## Limitations

There are certain known limitations within the application: 

- Interleaving anomalies: There is the case where two clients insert an entire word concurrently. The application will follow the mathematical properties of YATA, which may affect the meaning of the words. The final state could be an unpredictable output, which is a known limitation of CRDTs algorithms in general. The computational logic cannot guarantee words with a real meaning. 
- Tombstone accumulation: When a user deletes a character, the character gets deleted visually but maintains its logical CRDT structure. This is because other replicas could reference the deleted character for pending insert operations. The current mechanism is a logic deletion using a `bool` value to indicate which characters must be seen. This can lead to the accumulation of unseen characters and each operation that requires go through the entire list of characters can delay more and more as the garbage characters increase. The application currently does not have any garbage collection (GC) mechanism.
- Single point of failure: The application starts with the relay server that serves on port 8181. If the relay server gets disconnected all of the clients will be disconnected, the operations will remain locally until the relay starts again and the clients can communicate normally.
- Plain text only: The application does not support features such as bold, italic or underlined text. It only supports plain text and certain emojis.

---

## Future Work

There are certain implementations that could improve the project:

- Horizontal Scaling: The current deployment runs a single relay instance. Future work could explore horizontal scaling across multiple relay instances, including the synchronization and coordination mechanisms required to preserve CRDT convergence across replicas. Kubernetes could then be used to orchestrate these instances, provide service discovery, and manage their lifecycle.
- Observability: The system could expose metrics such as active WebSocket connections, document synchronization operations, update propagation latency, database activity, and error rates. Prometheus could be used for metrics collection, while Grafana could provide dashboards for monitoring the system under different workloads and network conditions.
- Garbage Collection (GC): The current implementation does not provide a garbage-collection mechanism for obsolete CRDT metadata. Over long-running editing sessions, deleted characters and historical operations may cause the document state to grow continuously. Future work could investigate safe garbage-collection strategies that reduce storage and memory usage without compromising convergence or consistency.
- Rich text: The current editor focuses on plain-text collaborative editing. Future versions could extend the CRDT model to represent formatting attributes such as bold text, italics, headings, lists, and other structured content. This introduces additional challenges because formatting operations must also converge consistently across concurrent edits.
- Awareness Protocol: The current implementation synchronizes document content but does not maintain ephemeral collaboration state such as cursor position, text selection, or user presence. Future work could introduce an awareness protocol that propagates this transient information among connected clients without storing it as part of the persistent CRDT document state. This would allow users to visualize collaborators' cursors and selections in real time while keeping awareness data separate from the replicated document model.

---

## References

