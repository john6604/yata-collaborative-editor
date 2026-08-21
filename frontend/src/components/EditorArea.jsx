import { forwardRef } from 'react';

const EditorArea = forwardRef(function EditorArea({
  displayName,
  documentId,
  onCursorChange,
  onInput,
  onCompositionEnd,
  content,
}, ref) {
  return (
    <section className="min-w-0 flex-1 p-5" aria-label="Document editor">
      <textarea
        className="h-full min-h-[620px] w-full resize-none rounded-lg border border-slate-300 bg-white p-6 font-mono text-base leading-7 text-slate-950 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-slate-900 focus:ring-2 focus:ring-slate-200"
        data-client-name={displayName}
        data-document-id={documentId}
        placeholder="Start writing..."
        ref={ref}
        onClick={onCursorChange}
        onChange={onInput}
        onCompositionEnd={onCompositionEnd}
        onKeyUp={onCursorChange}
        onSelect={onCursorChange}
        value={content}
        spellCheck="false"
      />
    </section>
  );
});

export default EditorArea;
