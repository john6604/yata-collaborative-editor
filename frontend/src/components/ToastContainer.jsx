const TONE_CLASSES = {
  info: 'border-sky-200 bg-sky-50 text-sky-950',
  success: 'border-emerald-200 bg-emerald-50 text-emerald-950',
  warning: 'border-amber-200 bg-amber-50 text-amber-950',
};

function ToastContainer({ toasts, onDismiss }) {
  return (
    <div className="pointer-events-none fixed bottom-5 right-5 z-50 flex w-full max-w-sm flex-col gap-3">
      {toasts.map((toast) => (
        <div
          className={`pointer-events-auto flex items-start justify-between gap-3 rounded-lg border px-4 py-3 text-sm shadow-sm transition ${TONE_CLASSES[toast.type] || TONE_CLASSES.info}`}
          key={toast.id}
          role="status"
        >
          <p className="leading-6">{toast.message}</p>
          <button
            aria-label="Dismiss notification"
            className="grid h-6 w-6 shrink-0 place-items-center rounded-md text-xs font-semibold transition hover:bg-black/5 focus:outline-none focus:ring-2 focus:ring-current"
            onClick={() => onDismiss(toast.id)}
            type="button"
          >
            x
          </button>
        </div>
      ))}
    </div>
  );
}

export default ToastContainer;
