import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import "@/styles.css";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("Workbench root element was not found.");
}

const root = createRoot(rootElement);

window.addEventListener("error", (event) => {
  renderBootError(event.error, event.message);
});

window.addEventListener("unhandledrejection", (event) => {
  renderBootError(event.reason, "Unhandled promise rejection during workbench boot.");
});

void bootstrap();

async function bootstrap() {
  try {
    const { App } = await import("@/App");
    root.render(
      <StrictMode>
        <App />
      </StrictMode>,
    );
  } catch (error) {
    renderBootError(error, "The workbench failed to initialize.");
  }
}

function renderBootError(error: unknown, fallbackMessage: string) {
  const message = error instanceof Error ? error.message : String(error || fallbackMessage);
  const stack = error instanceof Error && error.stack ? error.stack : "";

  root.render(
    <div className="min-h-screen bg-background px-6 py-10 text-foreground">
      <div className="mx-auto max-w-3xl rounded-3xl border border-danger/40 bg-card/95 p-6 shadow-sm">
        <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-danger">Workbench Boot Error</p>
        <h1 className="mt-3 text-2xl font-bold">The UI crashed before it could render.</h1>
        <p className="mt-3 text-sm leading-6 text-muted-foreground">{message || fallbackMessage}</p>
        {stack ? <pre className="code-block mt-5 whitespace-pre-wrap">{stack}</pre> : null}
      </div>
    </div>,
  );
}
