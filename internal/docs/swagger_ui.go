package docs

const SwaggerBeforeScript = `
const style = document.createElement('style');
style.textContent = [
  '.swagger-auth-modal .modal-ux-content > h3:not(.swagger-auth-title) { display: none; }',
  '.swagger-auth-modal .swagger-auth-title { font-size: 1.5em; font-weight: 600; margin: 0 0 16px; }',
  '.swagger-auth-modal .auth-container h4,',
  '.swagger-auth-modal .auth-container p,',
  '.swagger-auth-modal .auth-container .wrapper > span,',
  '.swagger-auth-modal .auth-container .wrapper > div:not(:has(input)) { display: none !important; }',
  '.swagger-auth-modal .auth-container label[for*="api_key"] { font-size: 0; }',
  '.swagger-auth-modal .auth-container label[for*="api_key"]::before { content: "token"; font-size: 14px; }',
].join('\n');
document.head.appendChild(style);
`

const SwaggerAfterScript = `
const authModalObserver = new MutationObserver(() => {
  document.querySelectorAll('.dialog-ux .modal-ux').forEach((modal) => {
    if (modal.dataset.authStyled) return;
    modal.dataset.authStyled = '1';
    modal.classList.add('swagger-auth-modal');

    const content = modal.querySelector('.modal-ux-content');
    if (!content || content.querySelector('.swagger-auth-title')) return;

    const title = document.createElement('h3');
    title.className = 'swagger-auth-title';
    title.textContent = 'Authorization';
    content.prepend(title);
  });
});
authModalObserver.observe(document.body, { childList: true, subtree: true });
`
