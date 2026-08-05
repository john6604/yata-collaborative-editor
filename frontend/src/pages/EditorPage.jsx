import { useMemo } from 'react';
import { useParams } from 'react-router-dom';
import EditorArea from '../components/EditorArea.jsx';
import Sidebar from '../components/Sidebar.jsx';
import StatusBar from '../components/StatusBar.jsx';
import TopBar from '../components/TopBar.jsx';

function EditorPage({
  characterCount,
  connectionStatus,
  cursorCol,
  cursorLine,
  displayName,
  documents,
  editorRef,
  isSidebarCollapsed,
  onToggleSidebar,
  pendingCount,
  users,
}) {
  const { documentId } = useParams();
  const document = useMemo(() => {
    return documents.find((currentDocument) => currentDocument.id === documentId);
  }, [documentId, documents]);

  return (
    <main className="flex min-h-screen flex-col bg-slate-100">
      <TopBar
        connectionStatus={connectionStatus}
        displayName={displayName}
        documentName={document?.name || 'Untitled Document'}
      />

      <div className="flex min-h-0 flex-1">
        <EditorArea ref={editorRef} displayName={displayName} documentId={documentId} />
        <Sidebar
          isCollapsed={isSidebarCollapsed}
          onToggle={onToggleSidebar}
          users={users}
        />
      </div>

      <StatusBar
        characterCount={characterCount}
        cursorCol={cursorCol}
        cursorLine={cursorLine}
        pendingCount={pendingCount}
      />
    </main>
  );
}

export default EditorPage;
