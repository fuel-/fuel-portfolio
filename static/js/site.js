// Hero typing, stat count-up, scroll reveal, and terminal overlay glue.
(function () {
  var reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // --- hero typing -------------------------------------------------------
  var cmd = document.getElementById("type-cmd");
  var output = document.querySelector("[data-hero-output]");
  function revealHero() {
    if (output) output.classList.remove("opacity-0");
  }
  if (cmd && output) {
    var text = "whoami";
    if (reduced) {
      // Under reduced-motion: set text directly (markup default "whoami" is
      // fine too, but we set it explicitly to be certain).
      cmd.textContent = text;
      revealHero();
    } else {
      // Clear the markup default ("whoami") so we can type it out fresh.
      cmd.textContent = "";
      var i = 0;
      setTimeout(function tick() {
        cmd.textContent = text.slice(0, ++i);
        if (i < text.length) setTimeout(tick, 90 + Math.random() * 70);
        else setTimeout(revealHero, 250);
      }, 500);
    }
  }

  // --- stat count-up -----------------------------------------------------
  // Animates the numeric prefix of strings like "99.9%", "30+", "15 yrs".
  var counters = document.querySelectorAll("[data-countup]");
  function parseStat(s) {
    var m = s.match(/^([\d.]+)(.*)$/);
    if (!m) return null;
    return { n: parseFloat(m[1]), suffix: m[2], decimals: (m[1].split(".")[1] || "").length };
  }
  if (!reduced && "IntersectionObserver" in window && counters.length) {
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        io.unobserve(entry.target);
        var p = parseStat(entry.target.textContent.trim());
        if (!p) return;
        var start = performance.now();
        (function step(now) {
          var t = Math.min(1, (now - start) / 900);
          var eased = 1 - Math.pow(1 - t, 3);
          entry.target.textContent = (p.n * eased).toFixed(p.decimals) + p.suffix;
          if (t < 1) requestAnimationFrame(step);
        })(start);
      });
    }, { threshold: 0.4 });
    counters.forEach(function (el) {
      var p = parseStat(el.textContent.trim());
      if (p) el.textContent = (0).toFixed(p.decimals) + p.suffix;
      io.observe(el);
    });
  }

  // --- scroll reveal -----------------------------------------------------
  var revealEls = document.querySelectorAll("[data-reveal]");
  if (reduced || !("IntersectionObserver" in window)) {
    revealEls.forEach(function (el) { el.classList.add("revealed"); });
  } else {
    var io2 = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) {
          e.target.classList.add("revealed");
          io2.unobserve(e.target);
        }
      });
    }, { threshold: 0.15 });
    revealEls.forEach(function (el) { io2.observe(el); });
  }

  // --- terminal glue -----------------------------------------------------
  // Server commands signal client behavior via the HX-Trigger header,
  // which htmx re-fires as a "term-action" DOM event.
  document.body.addEventListener("term-action", function (e) {
    var action = e.detail && e.detail.value;
    if (!action) return;
    if (action === "clear") {
      var out = document.getElementById("term-output");
      if (out) out.innerHTML = "";
    } else if (action === "exit") {
      window.dispatchEvent(new CustomEvent("terminal-close"));
    } else if (action.indexOf("goto:") === 0) {
      window.dispatchEvent(new CustomEvent("terminal-close"));
      var el = document.querySelector(action.slice(5));
      if (el) el.scrollIntoView({ behavior: reduced ? "auto" : "smooth" });
    } else if (action.indexOf("open:") === 0) {
      window.dispatchEvent(new CustomEvent("terminal-close"));
      var url = action.slice(5);
      var slug = url.split("/").pop();
      var btn = document.querySelector('[data-project-link="' + slug + '"]');
      if (btn) {
        btn.scrollIntoView({ behavior: reduced ? "auto" : "smooth", block: "center" });
        btn.click();
      } else {
        window.location.assign(url);
      }
    }
  });

  // Clear the input after each command and keep output pinned to bottom.
  document.body.addEventListener("htmx:afterRequest", function (e) {
    var form = document.getElementById("terminal-form");
    if (form && e.target === form) form.querySelector("input[name=cmd]").value = "";
  });
  document.body.addEventListener("htmx:afterSwap", function (e) {
    var target = e.detail && e.detail.target;
    if (target && target.id === "term-output") target.scrollTop = target.scrollHeight;
  });
})();
