import { useState } from 'react';
import { useHistory } from 'react-router-dom';

function DocumentListPage({ displayName, documents, onCreateDocument, onRenameDocument }) {
  const history = useHistory();
  const [names, setNames] = useState({});

  const handleCreateDocument = () => {
    const document = onCreateDocument();
    history.push(`/editor/${document.id}`);
  };

  const handleOpenDocument = (documentId) => {
    history.push(`/editor/${documentId}`);
  };

  return (
    <main className="min-h-screen bg-slate-100">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <h1 className="text-lg font-semibold tracking-normal text-slate-950">
            CRDT Collaborative Editor
          </h1>
          <div className="rounded-md border border-slate-200 px-3 py-2 text-sm font-medium text-slate-700">
            {displayName}
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-7xl px-6 py-8">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="text-2xl font-semibold tracking-normal text-slate-950">Documents</h2>
            <p className="mt-1 text-sm text-slate-600">Local session workspace</p>
          </div>
          <button
            className="rounded-md bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2"
            onClick={handleCreateDocument}
            type="button"
          >
            New Document
          </button>
        </div>

        {documents.length > 0 ? (
          <div className="mt-7 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {documents.map((document) => (
              <article
                className="cursor-pointer rounded-lg border border-slate-200 bg-white p-5 shadow-sm transition hover:border-slate-300 hover:shadow"
                key={document.id}
                onClick={() => handleOpenDocument(document.id)}
              >
                <label className="block" onClick={(event) => event.stopPropagation()}>
                  <span className="sr-only">Document name</span>
                  <input
                    className="w-full rounded-md border border-transparent bg-slate-50 px-3 py-2 text-base font-semibold text-slate-950 outline-none transition hover:border-slate-200 focus:border-slate-900 focus:bg-white focus:ring-2 focus:ring-slate-200"
                    onKeyDown={(event) => {
                      if (event.key === 'Enter') {
                        const name = names[document.id] ?? document.name;

                        if (name.trim() !== '') {
                          onRenameDocument(document.id, name);
                        }
                      }
                    }}
                    onChange={(event) => {
                      setNames((currentNames) => ({
                        ...currentNames,
                        [document.id]: event.target.value,
                      }));
                    }}
                    placeholder="Untitled Document"
                    type="text"
                    value={names[document.id] ?? document.name}
                  />
                </label>

                <dl className="mt-5 grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <dt className="font-medium text-slate-500">Last edited</dt>
                    <dd className="mt-1 text-slate-900">Just now</dd>
                  </div>
                  <div>
                    <dt className="font-medium text-slate-500">Active users</dt>
                    <dd className="mt-1 text-slate-900">0 users</dd>
                  </div>
                </dl>

                <button
                  className="mt-5 w-full rounded-md border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-300"
                  onClick={(event) => {
                    event.stopPropagation();
                    handleOpenDocument(document.id);
                  }}
                  type="button"
                >
                  Open
                </button>
              </article>
            ))}
          </div>
        ) : (
          <div className="mt-7 rounded-lg border border-dashed border-slate-300 bg-white p-10 text-center">
            <h3 className="text-base font-semibold text-slate-950">No documents yet</h3>
            <p className="mt-2 text-sm text-slate-600">Create one to enter the editor.</p>
          </div>
        )}
      </section>
    </main>
  );
}

export default DocumentListPage;
