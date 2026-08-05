import { useHistory } from 'react-router-dom';

function EntryPage({ displayName, onDisplayNameChange }) {
  const history = useHistory();

  const handleSubmit = (event) => {
    event.preventDefault();
    onDisplayNameChange(displayName.trim() || 'Guest');
    history.push('/documents');
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-100 px-6 py-10">
      <section className="w-full max-w-md rounded-lg border border-slate-200 bg-white p-8 shadow-sm">
        <div className="space-y-3">
          <h1 className="text-3xl font-semibold tracking-normal text-slate-950">
            CRDT Collaborative Editor
          </h1>
          <p className="text-base leading-7 text-slate-600">
            Real-time collaborative editing powered by CRDTs
          </p>
        </div>

        <form className="mt-8 space-y-5" onSubmit={handleSubmit}>
          <label className="block">
            <span className="text-sm font-medium text-slate-700">Display name</span>
            <input
              className="mt-2 block w-full rounded-md border border-slate-300 bg-white px-4 py-3 text-base text-slate-950 outline-none transition focus:border-slate-900 focus:ring-2 focus:ring-slate-200"
              onChange={(event) => onDisplayNameChange(event.target.value)}
              placeholder="Your name"
              type="text"
              value={displayName}
            />
          </label>

          <button
            className="w-full rounded-md bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2"
            type="submit"
          >
            Continue
          </button>
        </form>
      </section>
    </main>
  );
}

export default EntryPage;
