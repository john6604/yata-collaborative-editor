import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Redirect, Route, Switch } from 'react-router-dom';
import ToastContainer from './components/ToastContainer.jsx';
import DocumentListPage from './pages/DocumentListPage.jsx';
import EditorPage from './pages/EditorPage.jsx';
import EntryPage from './pages/EntryPage.jsx';

// === AVAILABLE STATE FOR INTEGRATION ===
// displayName / setDisplayName - string used later as client ID
// documents / setDocuments - [{id, name, createdAt}]
// users / setUsers - [{id, name, color}]
// toastRef.current.showToast(message, type) - show notification
// editorRef - ref to editor DOM element
// =======================================

// Implement an API to get documents
/*
initial value:
fecth(URL) -> this returns a Promise

Luego necesito convertir la respuesta a JSON

const response = await fetch(URL)
const json = await response.json()

El lugar para hacer un fetch es en un useEffect()

useState empezara como array vacio inicialmente hasta que se ejecute el fecth

se maneja errores asi:
if (!response.ok) {
    // manejar error
}

fetch al ser asincrono se deberia utilizar
await y async



*/

const createDocumentId = () => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }

  return `doc-${Date.now()}`;
};

function App() {
  const [displayName, setDisplayName] = useState(() => {
    if (sessionStorage.getItem("display_name") !== null) {
      return sessionStorage.getItem("display_name");
    }
    return "";
  });
  const [documents, setDocuments] = useState([]);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [toasts, setToasts] = useState([]);
  const editorRef = useRef(null);
  const toastRef = useRef(null);
  const toastTimersRef = useRef(new Map());

  // === WEBSOCKET INTEGRATION ===
  // WebSocket connection, message handling, and reconnection logic
  // will be implemented manually here.
  //
  // Expected integration:
  // - Connect to Go backend via WebSocket
  // - Update connectionStatus state on open/close/error
  // - Route incoming messages to appropriate handlers
  // =============================
  const initializeWebSocketIntegration = () => {};

  // === CRDT INPUT HANDLER ===
  // Keystroke capture, translation to CRDT insert/delete operations,
  // and cursor position tracking will be implemented manually here.
  //
  // Expected integration:
  // - Capture input events from EditorArea
  // - Translate each keystroke into a CRDT insert or delete operation
  // - Send operations through WebSocket
  // ==========================
  const handleCrdtEditorInput = () => {};

  // === REMOTE OPERATION RENDERER ===
  // Logic to apply incoming remote operations to the editor
  // (insert characters, delete characters) without disrupting
  // the local user's cursor position will be implemented manually here.
  //
  // Expected integration:
  // - Receive remote insert/delete operations
  // - Apply them to the editor content
  // - Preserve local cursor position during remote updates
  // =================================
  const applyRemoteOperationToEditor = () => {};

  // === SYNC PROTOCOL ===
  // sync_step1/sync_step2 exchange, initial document loading,
  // and delta integration will be implemented manually here.
  //
  // Expected integration:
  // - Send state vector + delete set on connection
  // - Receive and integrate delta from peers
  // - Update editor content with synced document state
  // =====================
  const runSyncProtocolIntegration = () => {};

  // === SESSION MANAGEMENT ===
  // Room joining, client ID assignment, join_ack handling,
  // and leave logic will be implemented manually here.
  //
  // Expected integration:
  // - Send join message with room and client ID
  // - Handle join_ack response
  // - Update user list on join/leave events
  // ==========================
  const handleSessionManagement = () => {};

  // === DOCUMENT PERSISTENCE ===
  // Document list fetching, creation, and deletion
  // via the Go backend will be implemented manually here.
  // ============================
  const handleDocumentPersistence = () => {};

  const dismissToast = useCallback((toastId) => {
    setToasts((currentToasts) => currentToasts.filter((toast) => toast.id !== toastId));

    const timer = toastTimersRef.current.get(toastId);
    if (timer) {
      clearTimeout(timer);
      toastTimersRef.current.delete(toastId);
    }
  }, []);

  const showToast = useCallback(
    (message, type = 'info') => {
      const toastId = createDocumentId();
      const toast = { id: toastId, message, type };

      setToasts((currentToasts) => [...currentToasts, toast]);
      const timer = setTimeout(() => dismissToast(toastId), 3000);
      toastTimersRef.current.set(toastId, timer);
    },
    [dismissToast],
  );

  useEffect(() => {
    toastRef.current = { showToast };

    return () => {
      toastRef.current = null;
    };
  }, [showToast]);

  useEffect(() => {
    return () => {
      toastTimersRef.current.forEach((timer) => clearTimeout(timer));
      toastTimersRef.current.clear();
    };
  }, []);

  const effectiveDisplayName = useMemo(() => {
    return displayName.trim() || 'Guest';
  }, [displayName]);

  // async functions to make HTTP requests
  
  const getDocuments = async() => {
    try {
      const response = await fetch("/api/documents");

      if (!response.ok) {
        throw new Error(`HTTP Error: ${response.status}`);
      }

      const data = await response.json();

      const documents = data.map((document) => {
        return {
          id: document.document_id,
          name: document.document_name,
          createdAt: document.created_at,
          updatedAt: document.updated_at,
        };
      });
      setDocuments(documents);
    } catch (err) {
      console.error("error obtaining documents.", err);
    }
  }

  const createDocument = useCallback(async () => {
    try {
      const response = await fetch("/api/documents", {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name: `Untitled Document ${documents.length + 1}`
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP Error: ${response.status}`)
      }

      const data = await response.json();
      
      const newDocument = {
        id: data.document_id,
        name: data.document_name,
        createdAt: data.created_at,
        updatedAt: data.updated_at,
      };

      setDocuments((currentDocuments) => [newDocument, ...currentDocuments]);
      return newDocument;
    } catch (err) {
      console.error("error creating document.", err);
    }
    
  }, [documents.length]);

  const renameDocument = useCallback(async (documentID, nextName) => {
    try {
      const response = await fetch(`/api/documents/${documentID}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name: nextName
        })
      });

      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }

      const data = await response.json();

      const documentRenamed = {
        id: data.document_id,
        name: data.document_name,
        createdAt: data.created_at,
        updatedAt: data.updated_at,
      }

      setDocuments((currentDocuments) =>
        currentDocuments.map((document) =>
          document.id === documentRenamed.id ? documentRenamed : document,
        ),
      );

    } catch (err) {
      console.error("could not rename the document.", err);
    }
  }, []);

  const toggleSidebar = useCallback(() => {
    setIsSidebarCollapsed((currentValue) => !currentValue);
  }, []);

  // useEffect to load status
  useEffect(() => {
    getDocuments();
  }, [])

  function updateDisplayName(name) {
    setDisplayName(name);
    sessionStorage.setItem("display_name", name);
  }

  return (
    <>
      <Switch>
        <Route
          exact
          path="/"
          render={() => (
            <EntryPage displayName={displayName} onDisplayNameChange={updateDisplayName} />
          )}
        />
        <Route
          path="/documents"
          render={() => (
            <DocumentListPage
              displayName={effectiveDisplayName}
              documents={documents}
              onCreateDocument={createDocument}
              onRenameDocument={renameDocument}
            />
          )}
        />
        <Route
          path="/editor/:documentId"
          render={() => (
            <EditorPage
              displayName={effectiveDisplayName}
              documents={documents}
              editorRef={editorRef}
              isSidebarCollapsed={isSidebarCollapsed}
              onToggleSidebar={toggleSidebar}
            />
          )}
        />
        <Redirect to="/" />
      </Switch>
      <ToastContainer toasts={toasts} onDismiss={dismissToast} />
    </>
  );
}

export default App;
