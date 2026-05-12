document.getElementById('connectBtn').addEventListener('click', async () => {
    const connectBtn = document.getElementById('connectBtn');
    const originalText = connectBtn.textContent; // Сохраняем исходный текст

    // get form data from UI
    const data = {
        endpoint: document.getElementById('endpoint').value.trim(),
        accessKeyId: document.getElementById('accessKeyId').value.trim(),
        secretKey: document.getElementById('secretKey').value,
        useSSL: document.getElementById('useSSL').checked,
        region: document.getElementById('region').value.trim()
    };

    // --- ЭТАП 1: Показываем "Подключение..." ВСЕГДА ---
    connectBtn.textContent = 'Подключение...';
    connectBtn.disabled = true;

    // --- ЭТАП 2: Валидация с задержкой для имитации процесса ---
    if (!data.endpoint || !data.accessKeyId || !data.secretKey) {
        const messageDiv = document.getElementById('message');
        messageDiv.innerHTML = '<p style="color:red;">Пожалуйста, заполните все обязательные поля!</p>';

        // Подсветка пустых полей
        ['endpoint', 'accessKeyId', 'secretKey'].forEach(id => {
            const input = document.getElementById(id);
            if (!input.value.trim()) {
                input.style.borderColor = 'var(--error)';
                setTimeout(() => input.style.borderColor = '', 2000);
            }
        });

        // Возвращаем кнопку в исходное состояние через небольшую задержку (для эффекта)
        setTimeout(() => {
            connectBtn.textContent = originalText;
            connectBtn.disabled = false;
        }, 800);

        return;
    }

    // --- ЭТАП 3: Отправка запроса на сервер ---
    try {
        const response = await fetch('/api/connect', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        const messageDiv = document.getElementById('message');

        if (response.ok) {
            // save S3 config to localStorage
            localStorage.setItem('s3config', JSON.stringify({
                endpoint: data.endpoint,
                accessKeyId: data.accessKeyId,
                secretKey: data.secretKey,
                useSSL: data.useSSL,
                region: data.region
            }));

            // show success message and redirect
            messageDiv.innerHTML = '<p style="color:green;">✓ Подключено! Перенаправление...</p>';
            
            setTimeout(() => {
                window.location.href = '/borr.html';
            }, 1000);

            // ❗ Не возвращаем кнопку — потому что будет редирект

        } else {
            const errorText = await response.text();
            messageDiv.innerHTML = `<p style="color:red;">Ошибка подключения: ${errorText}</p>`;
            
            // Возвращаем кнопку в исходное состояние
            connectBtn.textContent = originalText;
            connectBtn.disabled = false;
        }
    } catch (error) {
        const messageDiv = document.getElementById('message');
        messageDiv.innerHTML = `<p style="color:red;">Ошибка сети: ${error.message}</p>`;
        
        // Возвращаем кнопку даже при сетевой ошибке
        connectBtn.textContent = originalText;
        connectBtn.disabled = false;
    }
});

// Enter в полях → клик по кнопке
document.querySelectorAll('input').forEach(input => {
    input.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            document.getElementById('connectBtn').click();
        }
    });
});