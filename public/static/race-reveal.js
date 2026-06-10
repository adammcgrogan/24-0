(function () {
  'use strict';

  var RACE_DELAY_MS = 220;   // gap between each race result appearing
  var START_DELAY_MS = 600;  // pause before the first race reveals

  var cells = document.querySelectorAll('[data-race-cell]');
  if (!cells.length) return;

  var winsEl    = document.getElementById('live-wins');
  var lossesEl  = document.getElementById('live-losses');
  var scoreHero = document.getElementById('score-hero');
  var postRace  = document.getElementById('post-race');
  var tierEl    = document.getElementById('tier-display');
  var tierDesc  = document.getElementById('tier-desc');

  // Hide everything that gets revealed progressively.
  for (var i = 0; i < cells.length; i++) {
    cells[i].style.opacity = '0';
    cells[i].style.transform = 'scale(0.6)';
    cells[i].style.transition = 'none';
  }
  if (postRace)  postRace.style.display = 'none';
  if (tierEl)    tierEl.style.opacity   = '0';

  // Update the live score display.
  var wins = 0, losses = 0;

  function updateScore() {
    var total = wins + losses;
    if (winsEl)   winsEl.textContent   = wins;
    if (lossesEl) lossesEl.textContent = (24 - wins);

    // Pulse the score hero on each result.
    if (scoreHero) {
      scoreHero.style.transition = 'transform .12s cubic-bezier(.34,1.56,.64,1)';
      scoreHero.style.transform  = 'scale(1.04)';
      setTimeout(function () {
        scoreHero.style.transform = 'scale(1)';
      }, 130);
    }
  }

  var idx = 0;

  function revealNext() {
    if (idx >= cells.length) {
      onAllRevealed();
      return;
    }

    var cell = cells[idx];
    var won  = cell.dataset.won  === 'true';
    var dnf  = cell.dataset.dnf  === 'true';

    cell.style.transition = 'opacity .22s ease, transform .22s cubic-bezier(.16,1,.3,1)';
    cell.style.opacity    = '1';
    cell.style.transform  = 'scale(1)';

    if (won)       wins++;
    else if (!dnf) losses++;

    updateScore();

    idx++;
    setTimeout(revealNext, RACE_DELAY_MS);
  }

  function onAllRevealed() {
    // Reveal tier label + description with a fade.
    if (tierEl) {
      tierEl.style.transition = 'opacity .5s ease';
      tierEl.style.opacity = '1';
    }
    if (tierDesc) {
      tierDesc.style.opacity = '1';
    }
    // Reveal post-race section (stats, share, submit).
    if (postRace) {
      postRace.style.display = 'block';
      postRace.style.opacity = '0';
      postRace.style.transition = 'opacity .5s ease';
      // Force repaint before transitioning.
      postRace.getBoundingClientRect();
      postRace.style.opacity = '1';
    }
  }

  setTimeout(revealNext, START_DELAY_MS);
})();
