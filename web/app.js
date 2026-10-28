(function () {
  var low = document.getElementById("show-low");
  var issuesEl = document.getElementById("issues");
  var flowsEl = document.getElementById("flows");

  function el(tag, text, className) {
    var node = document.createElement(tag);
    if (className) node.className = className;
    if (text != null) node.textContent = text;
    return node;
  }

  function clear(node) {
    while (node.firstChild) node.removeChild(node.firstChild);
  }

  function when(ms) {
    if (typeof ms !== "number") return "";
    try {
      return new Date(ms).toISOString();
    } catch (err) {
      return String(ms);
    }
  }

  function meta(parent, label, value) {
    var p = el("p");
    p.appendChild(el("span", label + " ", "k"));
    p.appendChild(document.createTextNode(value == null ? "" : String(value)));
    parent.appendChild(p);
  }

  function renderIssues(data) {
    clear(issuesEl);
    var list = data && data.issues ? data.issues : [];
    if (!list.length) {
      issuesEl.appendChild(el("p", "No issues in the last 7 days.", "empty"));
      return;
    }
    var cards = el("div", null, "cards");
    list.forEach(function (issue) {
      var card = document.createElement("article");
      var top = el("div", null, "card-top");
      top.appendChild(el("p", issue.type || "", "type"));
      var score = el("div", null, "score");
      score.appendChild(el("b", String(issue.impactScore)));
      score.appendChild(el("span", "score"));
      top.appendChild(score);
      card.appendChild(top);

      var grid = el("div", null, "meta");
      meta(grid, "Route", issue.route);
      meta(grid, "Element", issue.elementSignature);
      meta(grid, "Flow", issue.flowName || "");
      meta(grid, "Users", issue.users);
      meta(grid, "Occurrences", issue.occurrences);
      var conf = el("p");
      conf.appendChild(el("span", "Confidence ", "k"));
      var pill = el("span", issue.confidence || "", "pill " + (issue.confidence || ""));
      conf.appendChild(pill);
      grid.appendChild(conf);
      meta(grid, "Release", issue.release || "");
      meta(grid, "Important", issue.important ? "yes" : "no");
      meta(grid, "First seen", when(issue.firstSeen));
      meta(grid, "Last seen", when(issue.lastSeen));
      card.appendChild(grid);

      card.appendChild(el("h3", "Observed"));
      card.appendChild(el("p", issue.observed || "", "body"));
      card.appendChild(el("h3", "Explanation"));
      card.appendChild(el("p", issue.explanation || "", "body"));
      card.appendChild(el("h3", "Proposed change"));
      card.appendChild(el("p", issue.suggestion || "", "body"));
      cards.appendChild(card);
    });
    issuesEl.appendChild(cards);
  }

  function renderFlows(data) {
    clear(flowsEl);
    var flows = data && data.flows ? data.flows : [];
    if (!flows.length) {
      flowsEl.appendChild(el("p", "No flows in the last 7 days.", "empty"));
      return;
    }
    var table = document.createElement("table");
    var thead = document.createElement("thead");
    var hr = document.createElement("tr");
    ["Name", "Important", "Attempts", "Completed", "Abandoned", "Open", "Steps"].forEach(function (label) {
      hr.appendChild(el("th", label));
    });
    thead.appendChild(hr);
    table.appendChild(thead);
    var body = document.createElement("tbody");
    flows.forEach(function (flow) {
      var tr = document.createElement("tr");
      tr.appendChild(el("td", flow.name || ""));
      tr.appendChild(el("td", flow.important ? "yes" : "no"));
      tr.appendChild(el("td", String(flow.attempts)));
      tr.appendChild(el("td", String(flow.completed)));
      tr.appendChild(el("td", String(flow.abandoned)));
      tr.appendChild(el("td", String(flow.open)));
      var steps = (flow.steps || []).map(function (step) {
        return step.step + " reached " + step.reached + ", abandoned " + step.abandoned;
      }).join("; ");
      tr.appendChild(el("td", steps));
      body.appendChild(tr);
    });
    table.appendChild(body);
    flowsEl.appendChild(table);
  }

  function load() {
    var api = window.HAWKEYE_API || "";
    var issuesURL = api + "/v1/issues?limit=100";
    if (low && low.checked) issuesURL += "&include=low";
    fetch(issuesURL, { cache: "no-store" })
      .then(function (res) { return res.json(); })
      .then(renderIssues)
      .catch(function () {});
    fetch(api + "/v1/flows", { cache: "no-store" })
      .then(function (res) { return res.json(); })
      .then(renderFlows)
      .catch(function () {});
  }

  if (low) low.addEventListener("change", load);
  load();
  setInterval(load, 10000);
})();
