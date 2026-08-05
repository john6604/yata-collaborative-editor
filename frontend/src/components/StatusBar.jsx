function StatusBar({ characterCount, cursorLine, cursorCol, pendingCount }) {
  return (
    <footer className="grid h-10 shrink-0 grid-cols-3 items-center border-t border-slate-200 bg-white px-5 text-xs font-medium text-slate-600">
      <div>{characterCount} characters</div>
      <div className="text-center">
        Line {cursorLine}, Col {cursorCol}
      </div>
      <div className="text-right">{pendingCount > 0 ? `${pendingCount} pending` : null}</div>
    </footer>
  );
}

export default StatusBar;
