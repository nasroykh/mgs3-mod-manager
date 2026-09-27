"use strict";

// The window for internal/gui. All text is set with textContent; nothing
// from the manager is parsed as HTML.

const $ = (id) => document.getElementById(id);
let api = null;
let rt = null;
let root = "";
let doctorOK = false;
let doctorDetails = "";
let overview = null;
let launchInfo = null;
let busy = false;

// ---------- small helpers ----------

function icon(name, extra) {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("class", "icon" + (extra ? " " + extra : ""));
  svg.setAttribute("aria-hidden", "true");
  const use = document.createElementNS("http://www.w3.org/2000/svg", "use");
  use.setAttribute("href", "icons.svg#i-" + name);
  svg.appendChild(use);
  return svg;
}

const toneIcon = { ok: "circle-check", warn: "alert-triangle", bad: "alert-circle", info: "info-circle", busy: "loader-2" };

function setBadge(el, text, tone) {
  el.replaceChildren();
  el.className = "badge";
  if (!text) return;
  if (tone && tone !== "info" && tone !== "busy") el.classList.add(tone);
  el.append(icon(toneIcon[tone || "info"], tone === "busy" ? "spin" : ""), document.createTextNode(text));
}

function toast(text, tone) {
  const t = document.createElement("div");
  t.className = "toast " + (tone || "ok");
  t.append(icon(toneIcon[tone || "ok"]), document.createTextNode(text));
  $("toasts").appendChild(t);
  setTimeout(() => t.remove(), 3500);
}

async function copy(text, button) {
  const ok = await rt.ClipboardSetText(text);
  const label = button.querySelector("span");
  const was = label.textContent;
  label.textContent = ok ? "Copied" : "Copy failed";
  setTimeout(() => { label.textContent = was; }, 1500);
}

function updateControls() {
  for (const el of $("app").querySelectorAll("button, select")) {
    el.disabled = busy || (!root && el.hasAttribute("data-needs-root"));
  }
  if (busy || !root) return;
  const state = overview ? overview.deltaState : "unknown";
  const kitStored = overview && overview.mods.some((m) => m.kit);
  $("uninstall").disabled = !kitStored;
  $("launch").disabled = !(doctorOK && launchInfo && launchInfo.available && selectedProfile() && selectedProfile().supported);
  const profiles = launchInfo && launchInfo.profiles.length > 0;
  $("profile").disabled = !profiles;
  $("profile-new").disabled = !profiles;
  $("profile-edit").disabled = !(profiles && selectedProfile());
  if (state === "installed") $("install").querySelector("span").textContent = "Reinstall / Verify";
  else $("install").querySelector("span").textContent = "Install / Update";
}

// ---------- activity dialog ----------

const dlg = {
  open: false,
  mode: "",        // confirm, running, done
  action: "",      // the service action whose events are shown
  steps: [],
  started: 0,
  timer: 0,
  runId: 0,
  returnKey: "",
};

const stepIcon = { waiting: "circle-dashed", running: "loader-2", done: "circle-check", skipped: "circle-minus", failed: "circle-x" };
const stepWord = { waiting: "", running: "Working…", done: "", skipped: "Nothing to do", failed: "Failed" };

function renderSteps() {
  const list = $("dlg-steps");
  list.replaceChildren();
  for (const s of dlg.steps) {
    const li = document.createElement("li");
    li.className = s.state;
    const label = document.createElement("span");
    label.textContent = s.label;
    const word = document.createElement("span");
    word.className = "state";
    word.textContent = stepWord[s.state] || "";
    li.append(icon(stepIcon[s.state] || "circle-dashed", s.state === "running" ? "spin" : ""), label, word);
    li.setAttribute("aria-label", s.label + (s.state === "waiting" ? "" : ": " + (stepWord[s.state] || s.state)));
    list.appendChild(li);
  }
  const total = dlg.steps.length;
  const finished = dlg.steps.filter((s) => s.state === "done" || s.state === "skipped").length;
  const bar = $("dlg-progress");
  bar.hidden = dlg.mode === "confirm" || total === 0;
  bar.classList.toggle("bad", dlg.steps.some((s) => s.state === "failed"));
  bar.setAttribute("aria-valuemax", String(total));
  bar.setAttribute("aria-valuenow", String(finished));
  $("dlg-bar").style.width = (total ? Math.round((finished / total) * 100) : 0) + "%";
  const running = list.querySelector("li.running");
  if (running) running.scrollIntoView({ block: "nearest" });
}

function setDialogHead(title, sub, tone, iconName, spinning) {
  $("dlg-title").textContent = title;
  $("dlg-sub").textContent = sub || "";
  const holder = $("dlg-icon");
  holder.className = "dlg-icon" + (tone ? " " + tone : "");
  holder.replaceChildren(icon(iconName, spinning ? "spin" : ""));
}

function setActions(buttons) {
  const box = $("dlg-actions");
  box.replaceChildren();
  for (const b of buttons) {
    if (b.spacer) {
      const gap = document.createElement("span");
      gap.className = "spacer";
      box.appendChild(gap);
      continue;
    }
    const el = document.createElement("button");
    el.type = "button";
    el.className = "btn " + (b.kind || "");
    if (b.icon) el.appendChild(icon(b.icon));
    const span = document.createElement("span");
    span.textContent = b.label;
    el.appendChild(span);
    el.addEventListener("click", b.onClick);
    box.appendChild(el);
  }
  return box.querySelectorAll("button");
}

// focusKey names the focused control so focus can come back to it after the
// page is re-rendered (mod switches are recreated by every refresh).
function focusKey(el) {
  if (!el || el === document.body) return "";
  if (el.dataset && el.dataset.mod) return "mod:" + el.dataset.mod;
  return el.id || "";
}

function restoreFocus() {
  const key = dlg.returnKey;
  if (!key || dlg.open || busy) return;
  dlg.returnKey = "";
  let el = key.startsWith("mod:") ? document.querySelector('.switch[data-mod="' + CSS.escape(key.slice(4)) + '"]') : $(key);
  if (!el || el.disabled) el = key.startsWith("mod:") ? $("mods-toggle") : $("recheck");
  if (el && !el.disabled) el.focus();
}

function openDialog(key) {
  if (!dlg.open) {
    dlg.returnKey = key === undefined ? focusKey(document.activeElement) : key;
    $("app").inert = true;
    $("overlay").hidden = false;
    dlg.open = true;
  }
  $("dlg-result").hidden = true;
  $("dlg-details").hidden = true;
  $("dlg-details").open = false;
  $("dlg-note").hidden = true;
  $("dlg-time").textContent = "";
  $("dlg-form").hidden = true;
  $("dlg-form").replaceChildren();
}

function closeDialog() {
  clearInterval(dlg.timer);
  dlg.open = false;
  dlg.mode = "";
  dlg.action = "";
  $("overlay").hidden = true;
  $("app").inert = false;
  restoreFocus(); // or at the end of the refresh that is still running
}

function tick() {
  const s = Math.floor((Date.now() - dlg.started) / 1000);
  $("dlg-time").textContent = Math.floor(s / 60) + ":" + String(s % 60).padStart(2, "0");
}

// run shows the activity dialog for one service call and resolves with its
// outcome. Short actions close the dialog by themselves when they succeed.
async function run(action, title, call, opts) {
  opts = opts || {};
  if (busy) {
    toast("Wait for the current check to finish.", "warn");
    return null;
  }
  const key = focusKey(document.activeElement);
  busy = true;
  updateControls();
  openDialog(dlg.open ? undefined : key);
  dlg.runId++;
  dlg.mode = "running";
  dlg.action = action;
  if (!opts.keepSteps) dlg.steps = [];
  renderSteps();
  setDialogHead(title, "Keep this window open until it finishes.", "", "loader-2", true);
  setActions([]);
  $("dialog").focus();
  dlg.started = Date.now();
  tick();
  dlg.timer = setInterval(tick, 500);
  let out;
  try {
    out = await call();
  } catch (e) {
    out = { ok: false, summary: "The window app failed: " + e, steps: [], details: String(e) };
  }
  clearInterval(dlg.timer);
  busy = false;
  dlg.mode = "done";
  showResult(title, out, opts);
  await refresh();
  return out;
}

function showResult(title, out, opts) {
  const ok = out && out.ok;
  setDialogHead(ok ? (opts.doneTitle || "Done") : "Stopped", ok ? title : (out.failed ? "At step: " + out.failed : title), ok ? "ok" : "bad", ok ? "circle-check" : "circle-x", false);
  const banner = $("dlg-result");
  banner.replaceChildren(icon(ok ? "circle-check" : "alert-triangle"), document.createTextNode(out.summary || ""));
  banner.className = "banner " + (ok ? "ok" : "bad");
  banner.hidden = !out.summary;
  $("dlg-pre").textContent = out.details || "";
  $("dlg-details").hidden = !out.details;
  $("dlg-note").hidden = true;
  announce(ok ? "Finished: " + (out.summary || "") : "Stopped: " + (out.summary || ""));
  const buttons = setActions([{ label: "Close", kind: ok ? "primary" : "", onClick: closeDialog }]);
  buttons[0].focus();
  if (ok && opts.autoClose) {
    const id = dlg.runId;
    setTimeout(() => {
      if (dlg.open && dlg.mode === "done" && dlg.runId === id) {
        closeDialog();
        toast(out.summary, "ok");
      }
    }, 900);
  }
}

function onProgress(e) {
  if (!dlg.open || dlg.mode !== "running" || e.action !== dlg.action) return;
  if (e.kind === "plan") {
    dlg.steps = e.steps.map((s) => ({ label: s.label, state: s.state }));
  } else if (e.kind === "step" && dlg.steps[e.index]) {
    const step = dlg.steps[e.index];
    step.state = e.state;
    if (e.state !== "waiting") announce(step.label + ": " + (stepWord[e.state] || e.state));
    if (e.action === "launch" && e.index === 1 && e.state === "running") {
      // The manager keeps its lock until it has seen the game process; tell
      // the user if that takes long.
      const id = dlg.runId;
      setTimeout(() => {
        if (dlg.runId === id && dlg.mode === "running" && dlg.steps[1] && dlg.steps[1].state === "running") {
          $("dlg-note").textContent = "Still waiting for the game to show up. The manager keeps the game folder locked until it has seen the game process, or until the game closes.";
          $("dlg-note").hidden = false;
        }
      }, 10000);
    }
  }
  renderSteps();
}

// announce reads progress to screen readers without re-reading the list.
function announce(text) {
  $("announce").textContent = text;
}

function onCloseRefused() {
  if (dlg.open) {
    const note = $("dlg-note");
    note.textContent = "Please wait: closing now could interrupt a change to the game files. The window can be closed when this finishes.";
    note.hidden = false;
    $("dialog").classList.remove("shake");
    void $("dialog").offsetWidth;
    $("dialog").classList.add("shake");
  } else {
    toast("Wait for the current action to finish before closing.", "warn");
  }
}

// Keyboard: keep Tab inside the dialog; Esc only cancels a confirmation or
// closes a finished action.
document.addEventListener("keydown", (e) => {
  if (!dlg.open) return;
  if (e.key === "Escape") {
    e.preventDefault();
    if (dlg.mode === "confirm" || dlg.mode === "done" || dlg.mode === "form") closeDialog();
    return;
  }
  if (e.key !== "Tab") return;
  const items = [...$("dialog").querySelectorAll("button, summary, input, select")].filter((el) => !el.disabled && el.offsetParent !== null);
  if (items.length === 0) { e.preventDefault(); return; }
  const first = items[0], last = items[items.length - 1];
  if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
  else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
  else if (!$("dialog").contains(document.activeElement)) { e.preventDefault(); first.focus(); }
});

// ---------- actions ----------

function install() {
  return run("install", "Installing Delta controls", () => api.Install(root), { doneTitle: "Delta controls installed" });
}

async function confirmUninstall() {
  if (busy || dlg.open) return;
  const key = focusKey(document.activeElement);
  // Hold the busy flag while the preview is read, so no action starts in
  // between and gets its dialog taken over by the confirmation.
  busy = true;
  updateControls();
  let preview;
  try {
    preview = await api.PlanUninstall(root);
  } catch (e) {
    preview = { steps: [], problem: { summary: "The window app failed: " + e, details: String(e) } };
  } finally {
    busy = false;
    updateControls();
  }
  openDialog(key);
  dlg.mode = "confirm";
  if (preview.problem) {
    setDialogHead("Cannot uninstall now", preview.problem.summary, "bad", "alert-triangle", false);
    dlg.steps = [];
    renderSteps();
    $("dlg-pre").textContent = preview.problem.details;
    $("dlg-details").hidden = false;
    setActions([{ label: "Close", kind: "primary", onClick: closeDialog }])[0].focus();
    return;
  }
  setDialogHead("Uninstall Delta controls?", "These steps will run. The original game files are put back.", "warn", "trash", false);
  dlg.steps = preview.steps.map((label) => ({ label, state: "waiting" }));
  renderSteps();
  if (preview.note) {
    $("dlg-note").textContent = preview.note;
    $("dlg-note").hidden = false;
  }
  const buttons = setActions([
    { label: "Cancel", onClick: closeDialog },
    { label: "Uninstall", kind: "destructive", icon: "trash", onClick: () => {
      $("dlg-note").hidden = true;
      run("uninstall", "Uninstalling Delta controls", () => api.Uninstall(root), { doneTitle: "Delta controls removed" });
    } },
  ]);
  buttons[0].focus();
}

function toggleMod(mod, on) {
  const verb = on ? "Turning on " : "Turning off ";
  return run(on ? "enable" : "disable", verb + mod.id, () => api.SetEnabled(root, mod.id, on), { autoClose: true, doneTitle: "Done" });
}

async function addPackage() {
  if (busy) return;
  const path = await api.ChoosePackage();
  if (path) run("add", "Adding a mod package", () => api.AddPackage(root, path), { autoClose: true, doneTitle: "Package stored" });
}

function verify() {
  return run("verify", "Verifying managed files", () => api.Verify(root), { autoClose: true, doneTitle: "Verify passed" });
}

function launchGame() {
  const p = selectedProfile();
  if (!p) return;
  return run("launch", "Starting the game", () => api.Launch(root, p.id), { autoClose: true, doneTitle: "Game started" });
}

// ---------- status ----------

function selectedProfile() {
  if (!launchInfo) return null;
  const id = $("profile").value;
  return launchInfo.profiles.find((p) => p.id === id) || null;
}

function renderLaunch() {
  const select = $("profile");
  const previous = select.value;
  select.replaceChildren();
  const info = launchInfo;
  for (const p of info ? info.profiles : []) {
    const opt = document.createElement("option");
    opt.value = p.id;
    opt.textContent = p.id + (p.default ? " (default)" : "") + (p.supported ? "" : " — not supported yet");
    select.appendChild(opt);
  }
  if (select.options.length === 0) {
    const opt = document.createElement("option");
    opt.value = "";
    opt.textContent = "No launch profile";
    select.appendChild(opt);
  }
  const def = info && info.profiles.find((p) => p.default);
  const wanted = preferProfile || previous;
  preferProfile = "";
  const kept = info && info.profiles.some((p) => p.id === wanted) ? wanted : "";
  if (kept) select.value = kept;
  else if (def) select.value = def.id;
  describeProfile();
}

// ---------- launch profiles ----------

let preferProfile = "";
let launchOptions = null;

function option(select, value, label) {
  const opt = document.createElement("option");
  opt.value = value;
  opt.textContent = label;
  select.appendChild(opt);
}

function field(labelText, control, wide) {
  const box = document.createElement("div");
  box.className = "field" + (wide ? " wide" : "");
  const label = document.createElement("label");
  label.textContent = labelText;
  label.htmlFor = control.id;
  box.append(label, control);
  return box;
}

// slug mirrors gui.ProfileID for the preview; the service decides.
function slug(name) {
  let out = "";
  let dash = false;
  for (const ch of name.trim().toLowerCase()) {
    if (/[a-z0-9._]/.test(ch)) { out += ch; dash = false; }
    else if ((ch === "-" || ch === " " || ch === "/") && out && !dash) { out += "-"; dash = true; }
  }
  out = out.replace(/^[-._]+|[-._]+$/g, "");
  return out.length > 64 ? out.slice(0, 64).replace(/[-._]+$/, "") : out;
}

async function refreshLaunch() {
  launchInfo = await api.LaunchInfo(root);
  renderLaunch();
}

// openProfileDialog creates a launch profile, or edits one (existing).
async function openProfileDialog(existing) {
  if (busy || dlg.open) return;
  const key = focusKey(document.activeElement);
  if (!launchOptions) launchOptions = await api.LaunchOptions();
  openDialog(key);
  dlg.mode = "form";
  dlg.runId++;
  dlg.steps = [];
  renderSteps();
  setDialogHead(existing ? "Edit launch profile" : "New launch profile", "Launch game starts the game with these settings, skipping the Master Collection menus.", "", "player-play", false);
  const form = $("dlg-form");
  form.hidden = false;

  const name = document.createElement("input");
  name.type = "text";
  name.id = "pf-name";
  name.maxLength = 64;
  name.autocomplete = "off";
  name.placeholder = "For example: Europe PS5";
  name.value = existing ? existing.id : "";
  const idHint = document.createElement("span");
  idHint.className = "muted small";
  const region = document.createElement("select");
  region.id = "pf-region";
  for (const r of launchOptions.regions) option(region, r.value, r.label);
  const language = document.createElement("select");
  language.id = "pf-language";
  const controller = document.createElement("select");
  controller.id = "pf-controller";
  for (const c of launchOptions.controllers) option(controller, c.value, c.label);
  const makeDefault = document.createElement("input");
  makeDefault.type = "checkbox";
  makeDefault.id = "pf-default";
  makeDefault.checked = existing ? existing.default : false;
  makeDefault.disabled = !!(existing && existing.default);
  const check = document.createElement("label");
  check.className = "check";
  check.htmlFor = "pf-default";
  check.append(makeDefault, document.createTextNode(existing && existing.default
    ? "This is the default profile (make another one the default to change it)"
    : "Make this the default profile"));
  const status = document.createElement("p");
  status.className = "callout";
  const error = document.createElement("p");
  error.className = "error";
  error.setAttribute("role", "alert");

  function fillLanguages(keep) {
    const r = launchOptions.regions.find((x) => x.value === region.value);
    language.replaceChildren();
    for (const l of r.languages) option(language, l.value, l.label);
    if (keep && r.languages.some((l) => l.value === keep)) language.value = keep;
  }
  function update() {
    const id = slug(name.value);
    idHint.textContent = id ? "Saved as: " + id : "Use letters or numbers in the name.";
    const tested = launchOptions.tested.some((t) => region.value === t.region && language.value === t.language && controller.value === t.controller);
    status.textContent = tested
      ? "Tested: this combination was played with the manager."
      : "Not tested yet: these are the launcher's own settings, but this combination was never run with the manager. If the game does not start as expected, use the Master Collection launcher and please report it.";
    status.className = "callout" + (tested ? " ok" : "");
    error.textContent = "";
  }
  region.value = existing ? existing.region : "us";
  fillLanguages(existing ? existing.language : "en");
  controller.value = existing ? existing.controller : "kbd";
  region.addEventListener("change", () => { fillLanguages(language.value); update(); });
  name.addEventListener("input", update);
  language.addEventListener("change", update);
  controller.addEventListener("change", update);

  const nameField = field("Name", name, true);
  nameField.appendChild(idHint);
  const checkField = document.createElement("div");
  checkField.className = "field wide";
  checkField.appendChild(check);
  form.append(nameField, field("Game region", region), field("Language", language), field("Button prompts", controller, true), status, checkField, error);
  update();

  function pending(on) {
    dlg.mode = on ? "saving" : "form";
    for (const el of $("dialog").querySelectorAll("button, input, select")) el.disabled = on || (el === makeDefault && !!(existing && existing.default));
  }
  async function call(fn) {
    if (dlg.mode !== "form") return null;
    pending(true);
    let r;
    try {
      r = await fn();
    } catch (e) {
      r = { ok: false, summary: "The window app failed: " + e };
    } finally {
      pending(false);
    }
    if (!r.ok) error.textContent = r.summary;
    return r;
  }
  async function save() {
    const r = await call(() => api.SaveProfile(root, existing ? existing.id : "", name.value, region.value, language.value, controller.value, makeDefault.checked));
    if (!r || !r.ok) return;
    preferProfile = r.id;
    closeDialog();
    toast(r.summary, "ok");
    await refreshLaunch();
  }
  let armedAt = 0;
  async function remove(e) {
    // A second click confirms, but not within half a second: a double-click
    // must not delete.
    const now = Date.now();
    if (!armedAt || now - armedAt < 500) {
      if (!armedAt) {
        armedAt = now;
        e.currentTarget.querySelector("span").textContent = "Click again to delete";
      }
      return;
    }
    const r = await call(() => api.DeleteProfile(root, existing.id));
    if (!r || !r.ok) return;
    closeDialog();
    toast(r.summary, "ok");
    await refreshLaunch();
  }
  const buttons = [];
  if (existing && launchInfo && launchInfo.profiles.length > 1) buttons.push({ label: "Delete", kind: "destructive-outline", icon: "trash", onClick: remove }, { spacer: true });
  buttons.push({ label: "Cancel", onClick: closeDialog }, { label: existing ? "Save" : "Create profile", kind: "primary", onClick: save });
  setActions(buttons);
  name.addEventListener("keydown", (e) => { if (e.key === "Enter") save(); });
  name.focus();
}

function describeProfile() {
  const p = selectedProfile();
  $("profile-desc").textContent = p ? p.description : "";
  const tag = $("profile-tested");
  tag.hidden = !p;
  if (p) {
    tag.textContent = p.tested ? "Tested" : "Not tested yet";
    tag.className = "tag" + (p.tested ? "" : " warn");
    tag.title = p.tested ? "This combination was played with the manager." : "Uses the launcher's own settings, but was never run with the manager. Please report whether it works.";
  }
  const notes = [];
  if (!root) notes.push("Choose the game folder first.");
  if (launchInfo && !launchInfo.available && launchInfo.reason) notes.push(launchInfo.reason);
  else if (p && !p.supported) notes.push(p.reason);
  if (launchInfo && launchInfo.steamCopy) notes.push("This is a Steam copy. Quick launch has not been tried on a Steam copy yet; if it does not start, start the game from Steam and please report it.");
  if (root && !doctorOK && launchInfo && launchInfo.available) notes.push("Fix the game folder check above first.");
  $("launch-note").textContent = notes.join(" ");
  $("launch-note").hidden = notes.length === 0;
  updateControls();
}

function renderMods(mods) {
  const body = $("mods");
  body.replaceChildren();
  $("mods-count").textContent = String(mods.length);
  $("mods-empty").hidden = mods.length > 0;
  body.parentElement.hidden = mods.length === 0;
  for (const mod of mods) {
    const tr = document.createElement("tr");
    const who = document.createElement("td");
    const id = document.createElement("span");
    id.className = "mod-id";
    id.textContent = mod.id;
    who.appendChild(id);
    if (mod.kit) {
      const tag = document.createElement("span");
      tag.className = "tag";
      tag.textContent = "Delta kit";
      who.appendChild(tag);
    }
    if (mod.name && mod.name !== mod.id) {
      const name = document.createElement("span");
      name.className = "mod-name";
      name.textContent = mod.name;
      who.appendChild(name);
    }
    const version = document.createElement("td");
    version.className = "mono small";
    version.textContent = mod.version;
    const on = document.createElement("td");
    on.className = "right";
    const sw = document.createElement("button");
    sw.type = "button";
    sw.className = "switch";
    sw.setAttribute("role", "switch");
    sw.dataset.mod = mod.id;
    sw.setAttribute("aria-checked", mod.enabled ? "true" : "false");
    sw.setAttribute("aria-label", mod.id + (mod.enabled ? " is on" : " is off"));
    sw.title = mod.enabled ? "Turn off" : "Turn on";
    sw.addEventListener("click", () => toggleMod(mod, !mod.enabled));
    on.appendChild(sw);
    tr.append(who, version, on);
    body.appendChild(tr);
  }
  updateControls(); // new switches follow the busy state too
}

async function refresh() {
  const wasBusy = busy;
  busy = true;
  updateControls();
  try {
    $("copy").hidden = true;
    for (const card of document.querySelectorAll(".grid .card, #mods-title")) card.closest(".card").classList.toggle("dim", !root);
    if (!root) {
      doctorOK = false;
      overview = null;
      launchInfo = null;
      setBadge($("doctor"), "Choose your MGS3 folder: the one with METAL GEAR SOLID3.exe.", "warn");
      setBadge($("delta"), "Choose the game folder first.", "info");
      $("kit").textContent = "";
      renderMods([]);
      renderLaunch();
      return;
    }
    setBadge($("doctor"), "Checking the game folder…", "busy");
    setBadge($("delta"), "Reading installed mods…", "busy");
    const check = await api.Doctor(root);
    doctorOK = check.ok;
    doctorDetails = check.details;
    setBadge($("doctor"), check.summary, check.ok ? "ok" : "bad");
    $("copy").hidden = false;
    overview = await api.Status(root);
    const tone = { installed: "ok", missing: "info", partial: "warn", outdated: "warn", unknown: "bad" }[overview.deltaState];
    let line = overview.delta;
    if (overview.problem) line += " — " + overview.problem.summary;
    setBadge($("delta"), line, overview.problem ? "bad" : tone);
    $("kit").textContent = overview.kit;
    if (overview.problem) doctorDetails += "\nstatus: " + overview.problem.details;
    renderMods(overview.mods || []);
    launchInfo = await api.LaunchInfo(root);
    renderLaunch();
  } catch (e) {
    setBadge($("doctor"), "The window app failed: " + e, "bad");
  } finally {
    busy = wasBusy;
    updateControls();
    restoreFocus();
  }
}

// ---------- game folder ----------

function fillFolders(detection) {
  const select = $("folder");
  select.replaceChildren();
  const folders = detection.folders || [];
  if (folders.length === 0 || !detection.selected) {
    const opt = document.createElement("option");
    opt.value = "";
    opt.textContent = folders.length === 0 ? "No game folder found in your Steam libraries — use Browse…" : "Several game folders found — choose one";
    select.appendChild(opt);
  }
  for (const folder of folders) {
    const opt = document.createElement("option");
    opt.value = folder;
    opt.textContent = folder;
    opt.title = folder;
    select.appendChild(opt);
  }
  select.value = detection.selected || "";
  select.title = select.value;
  $("folder-note").hidden = !detection.note;
  $("folder-note").textContent = detection.note || "";
}

async function selectRoot(folder) {
  root = folder || "";
  $("folder").title = root;
  if (root) {
    const err = await api.Remember(root);
    if (err) toast("The folder could not be remembered: " + err, "warn");
  }
  await refresh();
}

async function browse() {
  if (busy) return;
  const dir = await api.ChooseFolder();
  if (!dir) return;
  const select = $("folder");
  const empty = [...select.options].find((o) => o.value === "");
  if (empty) empty.remove();
  if (![...select.options].some((o) => o.value === dir)) {
    const opt = document.createElement("option");
    opt.value = dir;
    opt.textContent = dir;
    opt.title = dir;
    select.appendChild(opt);
  }
  select.value = dir;
  await selectRoot(dir);
}

// ---------- start ----------

async function start() {
  api = window.go.main.App;
  rt = window.runtime;
  rt.EventsOn("progress", onProgress);
  rt.EventsOn("close-refused", onCloseRefused);
  $("dialog").tabIndex = -1;
  $("version").textContent = "v" + (await api.Version());
  $("folder").addEventListener("change", (e) => selectRoot(e.target.value));
  $("browse").addEventListener("click", browse);
  $("recheck").addEventListener("click", () => refresh());
  $("copy").addEventListener("click", () => copy("MGS3 Mod Manager " + $("version").textContent + "\n" + doctorDetails, $("copy")));
  $("dlg-copy").addEventListener("click", () => copy($("dlg-pre").textContent, $("dlg-copy")));
  $("install").addEventListener("click", install);
  $("uninstall").addEventListener("click", confirmUninstall);
  $("verify").addEventListener("click", verify);
  $("add").addEventListener("click", addPackage);
  $("launch").addEventListener("click", launchGame);
  $("profile").addEventListener("change", describeProfile);
  $("profile-new").addEventListener("click", () => openProfileDialog(null));
  $("profile-edit").addEventListener("click", () => openProfileDialog(selectedProfile()));
  $("open").addEventListener("click", async () => {
    const err = await api.OpenFolder(root);
    if (err) toast(err, "bad");
  });
  $("mods-toggle").addEventListener("click", () => {
    const open = $("mods-toggle").getAttribute("aria-expanded") !== "true";
    $("mods-toggle").setAttribute("aria-expanded", open ? "true" : "false");
    $("mods-body").hidden = !open;
  });
  setBadge($("doctor"), "Looking for the game…", "busy");
  const detection = await api.Detect();
  fillFolders(detection);
  await selectRoot(detection.selected);
}

function waitForBridge() {
  if (window.go && window.go.main && window.go.main.App && window.runtime) {
    start().catch((e) => setBadge($("doctor"), "The window app failed to start: " + e, "bad"));
  } else {
    setTimeout(waitForBridge, 20);
  }
}

document.addEventListener("DOMContentLoaded", waitForBridge);
