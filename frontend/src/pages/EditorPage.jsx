import { useMemo } from 'react';
import { useParams } from 'react-router-dom';
import EditorArea from '../components/EditorArea.jsx';
import Sidebar from '../components/Sidebar.jsx';
import StatusBar from '../components/StatusBar.jsx';
import TopBar from '../components/TopBar.jsx';
import { useCollaborativeDocument } from '../hooks/useCollaborativeDocument.js';

function EditorPage({
  displayName,
  documents,
  editorRef,
  isSidebarCollapsed,
  onToggleSidebar,
}) {
  const { documentId } = useParams();
  const {
    characterCount,
    connectionStatus,
    content,
    cursorCol,
    cursorLine,
    handleCompositionEnd,
    handleCursorChange,
    handleInput,
    pendingCount,
    users,
  } = useCollaborativeDocument(documentId, displayName, editorRef);
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
        <EditorArea
          ref={editorRef}
          content={content}
          displayName={displayName}
          documentId={documentId}
          onCompositionEnd={handleCompositionEnd}
          onCursorChange={handleCursorChange}
          onInput={handleInput}
        />
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
