import { Link } from 'react-router-dom';

const STATUS_STYLES = {
  online: {
    dot: 'bg-emerald-500',
    label: 'Connected',
  },
  syncing: {
    dot: 'bg-amber-400',
    label: 'Syncing...',
  },
  offline: {
    dot: 'bg-red-500',
    label: 'Offline',
  },
};

function TopBar({ connectionStatus, displayName, documentName }) {
  const status = STATUS_STYLES[connectionStatus] || STATUS_STYLES.offline;

  return (
    <header className="flex h-16 shrink-0 items-center justify-between border-b border-slate-200 bg-white px-5">
      <div className="flex min-w-0 items-center gap-4">
        <Link
          className="rounded-md border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-300"
          to="/documents"
        >
          Back to documents
        </Link>
        <div className="min-w-0">
          <h1 className="truncate text-base font-semibold tracking-normal text-slate-950">
            {documentName}
          </h1>
          <p className="mt-0.5 truncate text-xs font-medium text-slate-500">Client: {displayName}</p>
        </div>
      </div>

      <div className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm font-medium text-slate-700">
        <span className={`h-2.5 w-2.5 rounded-full ${status.dot}`} aria-hidden="true" />
        <span>{status.label}</span>
      </div>
    </header>
  );
}

export default TopBar;
