import React from 'react';

interface SyntaxHighlighterProps {
  code: string;
  language: string;
}

// A simple tokenizer for basic highlighting to avoid heavy dependencies in this demo
const SimpleHighlighter: React.FC<SyntaxHighlighterProps> = ({ code, language }) => {
  
  const renderTokens = () => {
    const lines = code.split('\n');
    return lines.map((line, i) => {
      // Very basic heuristic styling based on common config patterns
      let processedLine = line;
      
      // Comments
      if (line.trim().startsWith('#') || line.trim().startsWith('//')) {
        return <div key={i} className="text-slate-500 italic">{line}</div>;
      }

      // Keys in YAML/JSON (before colon)
      const keyMatch = line.match(/^(\s*)([\w\d_]+):/);
      if (keyMatch) {
        const [full, space, key] = keyMatch;
        const rest = line.slice(full.length);
        return (
          <div key={i}>
            <span className="whitespace-pre">{space}</span>
            <span className="text-sky-400 font-semibold">{key}:</span>
            <span className="text-emerald-300">{rest}</span>
          </div>
        );
      }

      // Dockerfile Instructions
      const dockerMatch = line.match(/^([A-Z]+)\s+(.*)/);
      if (language.toUpperCase() === 'DOCKERFILE' && dockerMatch) {
         const [_, instruction, args] = dockerMatch;
         return (
             <div key={i}>
                 <span className="text-purple-400 font-bold">{instruction}</span>
                 <span className="text-slate-300"> {args}</span>
             </div>
         )
      }

      // Default
      return <div key={i} className="text-slate-300 whitespace-pre">{line}</div>;
    });
  };

  return (
    <pre className="font-mono text-sm leading-6">
      {renderTokens()}
    </pre>
  );
};

export default SimpleHighlighter;
