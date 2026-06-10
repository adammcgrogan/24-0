(function () {
  'use strict';

  var RACE_DELAY_MS  = 280;  // gap between each race
  var START_DELAY_MS = 700;  // pause before the first race

  var cells = document.querySelectorAll('[data-race-cell]');
  if (!cells.length) return;

  var counterEl = document.getElementById('race-counter');
  var winsEl    = document.getElementById('live-wins');
  var lossesEl  = document.getElementById('live-losses');
  var scoreHero = document.getElementById('score-hero');
  var postRace  = document.getElementById('post-race');
  var tierEl    = document.getElementById('tier-display');

  var wins = 0, losses = 0, idx = 0;
  var total = cells.length;

  function updateScore(won, dnf) {
    if (won)       wins++;
    else if (!dnf) losses++;

    if (winsEl)   winsEl.textContent   = wins;
    if (lossesEl) lossesEl.textContent = losses;

    if (counterEl) {
      counterEl.textContent = 'RACE ' + (idx) + ' / ' + total;
    }

    if (scoreHero) {
      scoreHero.style.transition = 'transform .12s cubic-bezier(.34,1.56,.64,1)';
      scoreHero.style.transform  = 'scale(1.06)';
      setTimeout(function () {
        scoreHero.style.transform = 'scale(1)';
      }, 130);
    }
  }

  function revealNext() {
    if (idx >= total) {
      onAllRevealed();
      return;
    }

    var cell = cells[idx];
    var won  = cell.dataset.won === 'true';
    var dnf  = cell.dataset.dnf === 'true';

    // Add result class to the dot before revealing so colours appear instantly
    var dot = cell.querySelector('.race-dot');
    if (dot) {
      if (won)       dot.classList.add('win');
      else if (dnf)  dot.classList.add('dnf');
      else           dot.classList.add('loss');
    }

    // Trigger the CSS transition defined in base.html
    cell.classList.add('revealed');

    idx++;
    updateScore(won, dnf);

    setTimeout(revealNext, RACE_DELAY_MS);
  }

  function onAllRevealed() {
    if (counterEl) counterEl.textContent = 'SEASON COMPLETE';

    // tier-display has its own CSS transition covering opacity + transform
    if (tierEl) {
      tierEl.style.opacity   = '1';
      tierEl.style.transform = 'translateY(0)';
    }

    // Fade in the post-race section
    if (postRace) {
      postRace.style.display  = 'block';
      postRace.style.opacity  = '0';
      postRace.style.transition = 'opacity .5s ease';
      postRace.getBoundingClientRect(); // force repaint
      postRace.style.opacity  = '1';
    }
  }

  setTimeout(revealNext, START_DELAY_MS);
})();
