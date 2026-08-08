const boxes = document.querySelector("#boxes");
const output = document.querySelector("#output");
const status = document.querySelector("#status");
const refresh = document.querySelector("#refresh");
const form = document.querySelector("#create-form");
const nameInput = document.querySelector("#name");
const main = document.querySelector("main");
let busy = false;

function backend() {
  const api = window.go?.main?.App;
  if (!api) {
    throw new Error("Desktop API is unavailable. Start this page through Wails.");
  }
  return api;
}

function setStatus(message, failed = false) {
  status.textContent = message;
  status.classList.toggle("failed", failed);
}

function setBusy(isBusy) {
  busy = isBusy;
  main.setAttribute("aria-busy", String(isBusy));
  document.querySelectorAll("button, input").forEach((control) => {
    control.disabled = isBusy;
  });
}

function showOutput(text) {
  output.textContent = text || "Command completed without output.";
}

async function invoke(label, action) {
  if (busy) return false;

  setBusy(true);
  setStatus(`${label} in progress…`);
  try {
    const result = await action();
    showOutput(result);
    setStatus(`${label} complete`);
    await loadBoxes();
    return true;
  } catch (error) {
    showOutput(error?.message || String(error));
    setStatus(`${label} failed`, true);
    return false;
  } finally {
    setBusy(false);
  }
}

function actionButton(label, handler, danger = false) {
  const button = document.createElement("button");
  button.textContent = label;
  button.className = danger ? "danger" : "secondary";
  button.disabled = busy;
  button.addEventListener("click", handler);
  return button;
}

async function deleteBox(name) {
  const confirmation = window.prompt(
    `Delete box "${name}"? This permanently purges its persistent data.\n\nType the box name to confirm.`,
  );
  if (confirmation !== name) {
    setStatus(`Delete "${name}" cancelled`);
    return;
  }
  await invoke("Delete", () => backend().Delete(name));
}

async function setupWithDefaults(name) {
  const confirmed = window.confirm(
    `Set up "${name}" with CLI defaults?\n\nThis creates or recreates the box using the CLI defaults.`,
  );
  if (!confirmed) {
    setStatus(`Set up "${name}" cancelled`);
    return;
  }
  await invoke("Set up with defaults", () => backend().Setup(name));
}

function renderBoxes(text) {
  const lines = text.trim() ? text.trim().split("\n") : [];
  boxes.replaceChildren();
  if (lines.length === 0) {
    boxes.textContent = "No boxes yet. Create one to get started.";
    boxes.classList.add("empty");
    return;
  }
  boxes.classList.remove("empty");
  for (const line of lines) {
    const [name, state, image] = line.split("\t");
    const card = document.createElement("article");
    const details = document.createElement("div");
    const title = document.createElement("h3");
    const description = document.createElement("p");
    const actions = document.createElement("div");
    card.className = "box";
    actions.className = "actions";
    title.textContent = name;
    description.textContent = `${state} · ${image}`;
    details.append(title, description);
    card.append(details, actions);
    actions.append(
      actionButton("Inspect", () => invoke("Inspect", () => backend().Inspect(name))),
      actionButton("Set up with defaults", () => setupWithDefaults(name)),
      actionButton("Stop", () => invoke("Stop", () => backend().Stop(name))),
      actionButton("Delete", () => deleteBox(name), true),
    );
    boxes.append(card);
  }
}

async function loadBoxes() {
  try {
    const result = await backend().List();
    renderBoxes(result);
    return true;
  } catch (error) {
    showOutput(error?.message || String(error));
    setStatus("List failed", true);
    return false;
  }
}

refresh.addEventListener("click", () => invoke("Refresh", () => backend().List()));
form.addEventListener("submit", (event) => {
  event.preventDefault();
  const name = nameInput.value.trim();
  invoke("Create", () => backend().Create(name)).then((created) => {
    if (created) nameInput.value = "";
  });
});

async function initialize() {
  setBusy(true);
  setStatus("Loading boxes…");
  if (await loadBoxes()) setStatus("Ready");
  setBusy(false);
}

initialize();
