// Phosphor dot-grid background for the hero. A soft green glow drifts
// across a grid of dots. Renders a single static frame when the user
// prefers reduced motion.
(function () {
  var canvas = document.getElementById("grid-bg");
  if (!canvas) return;
  var ctx = canvas.getContext("2d");
  var reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var dpr = window.devicePixelRatio || 1;
  var w, h, t = 0;

  function resize() {
    w = canvas.width = canvas.offsetWidth * dpr;
    h = canvas.height = canvas.offsetHeight * dpr;
  }
  window.addEventListener("resize", resize);
  resize();

  var GAP = 28 * dpr;

  function frame() {
    ctx.clearRect(0, 0, w, h);
    var cx = w / 2 + Math.cos(t / 600) * w * 0.25;
    var cy = h * 0.4 + Math.sin(t / 800) * h * 0.15;
    for (var x = GAP / 2; x < w; x += GAP) {
      for (var y = GAP / 2; y < h; y += GAP) {
        var d = Math.hypot(x - cx, y - cy);
        var a = Math.max(0, 1 - d / (w * 0.45));
        if (a <= 0.05) continue;
        ctx.fillStyle = "rgba(74, 222, 128, " + (a * 0.35).toFixed(3) + ")";
        ctx.beginPath();
        ctx.arc(x, y, 1.1 * dpr, 0, Math.PI * 2);
        ctx.fill();
      }
    }
    t++;
    if (!reduced) requestAnimationFrame(frame);
  }
  frame();
})();
