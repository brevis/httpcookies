(() => {
    'use strict';

    const escapeHtml = unsafe => unsafe
         .replace(/&/g, "&amp;")
         .replace(/</g, "&lt;")
         .replace(/>/g, "&gt;")
         .replace(/"/g, "&quot;")
         .replace(/'/g, "&#039;");

    window.addEventListener('DOMContentLoaded', () => {
        const state = {
            method: 'GET',
            headersDefault: 'User-Agent: HttpCookies.info/1.0',
            moreOptions: false,
        };

        const form = document.getElementById('get_cookies');
        const submitButton = form.querySelector('button[type="submit"]');
        const urlInput = form.querySelector('.url_input');
        const methodButton = document.querySelector('.method_button');
        const methodItemsButtons = document.querySelectorAll('a.method_item');
        const resultWrapper = document.getElementById('result');
        const moreOprionsWrapper = document.querySelector('.more_options');
        const headersInput = moreOprionsWrapper.querySelector('.headers_input');
        const bodyInput = moreOprionsWrapper.querySelector('.body_input');
        const moreOprionsButtons = document.querySelectorAll('a.more_options_item');
        const bodyNotice = document.querySelector('.body_notice');
        
        const applyBodyNotice = () => {
            if (['GET', 'HEAD', 'OPTIONS', 'CONNECT'].indexOf(state.method) !== -1) {
                bodyNotice.classList.remove('hidden');
            } else {
                bodyNotice.classList.add('hidden');
            }
        };

        const onMethodItemClick = event => {
            const method = event.currentTarget.innerHTML;

            state.method = methodButton.innerHTML = method;

            methodItemsButtons.forEach(el => {
                el.classList.remove('hidden');
            });
            event.currentTarget.classList.add('hidden');

            applyBodyNotice();

            event.preventDefault();
            return false;
        };

        const onMoreOptionsClick = event => {
            state.moreOptions = !state.moreOptions;
            if (state.moreOptions) {
                moreOprionsWrapper.classList.remove('hidden');
            } else {
                moreOprionsWrapper.classList.add('hidden');
                headersInput.value = state.headersDefault;
                bodyInput.value = '';
            }            

            applyBodyNotice();

            event.preventDefault();
            return false;
        };

        
        methodItemsButtons.forEach(el => {
            el.addEventListener('click', onMethodItemClick);
        });

        moreOprionsButtons.forEach(el => {
            el.addEventListener('click', onMoreOptionsClick);
        });

        form.addEventListener('submit', event => {
            const body = new FormData();
            body.append('method', state.method);
            body.append('url', urlInput.value);
            body.append('headers', headersInput.value);
            body.append('body', bodyInput.value);

            result.innerHTML = '';
            submitButton.disabled = true;
            fetch(form.action, {
                method: 'POST',
                body: body
            })
            .then(response => response.json())
            .then(result => {
                if (result.error) {
                    resultWrapper.innerHTML = `<div class="alert alert-danger result_error" role="alert">${result.error}</div>`;    
                } else {
                    let statusHtml = '<table class="table"><tr><td><span class="badge text-bg-light">Status</span></td><td><code>' + result.status + '</code></td></tr></table>';
                    let cokiesHtml = '<p>No cookies were found in the response</p>';
                    if (result.cookies && result.cookies.length > 0) {
                        cokiesHtml = '<h4>Cookies</h4><div class="table-responsive"><table class="table"><thead><tr><th scope="col">Name</th><th scope="col">Value</th></tr></thead><tbody class="table-group-divider">';
                        result.cookies.forEach(c => {
                            cokiesHtml += '<tr><td><span class="badge text-bg-light">' + escapeHtml(c.name) + '</span></td><td><code>' + escapeHtml(c.value) + '</code></td>';
                        });
                        cokiesHtml += '</tbody></table></div>';
                    }
                    let headersHtml = '';
                    if (result.headers && result.headers.length > 0) {
                        headersHtml = '<a href="#" class="link-secondary show_headers">Show headers</a><div class="readers_result hidden"><h4>Headers</h4><div class="table-responsive"><table class="table"><thead><tr><th scope="col">Name</th><th scope="col">Value</th></tr></thead><tbody class="table-group-divider">';
                        result.headers.forEach(c => {
                            headersHtml += '<tr><td><span class="badge text-bg-light">' + escapeHtml(c.name) + '</span></td><td>';
                            c.values.forEach(v => {
                                headersHtml += '<div><code>' + escapeHtml(v) + '</code></div>';
                            });
                            headersHtml += '</td></tr>';
                        });
                        headersHtml += '</tbody></table></div>';
                    }
                    resultWrapper.innerHTML = `<div class="succes_result alert alert-light">${statusHtml}${cokiesHtml}${headersHtml}</div>`;
                    const showHeadersLink = document.querySelector('.show_headers');
                    if (showHeadersLink) {
                        showHeadersLink.addEventListener('click', event => {
                            document.querySelector('.readers_result').classList.remove('hidden');
                            event.currentTarget.remove();

                            event.preventDefault();
                            return false;
                        });
                    }
                }                
                submitButton.disabled = false;
            })
            .catch(error => {
                if (!error) {
                    error = 'Something went wrong :(';
                }
                resultWrapper.innerHTML = `<div class="alert alert-danger" role="alert">${error}</div>`;
                submitButton.disabled = false;
            });

            event.preventDefault();
            return false;
        });
    });
})();