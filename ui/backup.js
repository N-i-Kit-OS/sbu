const s3Config = JSON.parse(localStorage.getItem('s3config') || '{}');

if (!s3Config.endpoint || !s3Config.accessKeyId || !s3Config.secretKey ) {
    const messageDiv = document.getElementById('message');
    if (messageDiv) {
        messageDiv.innerHTML = '<p style="color:red;">Ошибка: сначала выполните подключение на главной странице.</p>';
    }
    document.getElementById('backupBtn').disabled = true;
}

document.getElementById('backupBtn').addEventListener('click', async () => {
    const source = document.getElementById('backupSourcePath').value.trim();
    const backupName = document.getElementById('backupName').value.trim();
    const bucketName = document.getElementById('backupBucket').value.trim();

    // Исправлено: используем правильный ID для сообщения
    const messageDiv = document.getElementById('backupMessage');
    
    if (!source) {
        messageDiv.innerHTML = '<p style="color:red;">Укажите путь к папке или файлу!</p>';
        return;
    }

    if (!backupName) {
        messageDiv.innerHTML = '<p style="color:red;">Укажите имя бэкапа!</p>';
        return;
    }

    if (!bucketName) {
        messageDiv.innerHTML = '<p style="color:red;">Укажите название bucket!</p>';
        return;
    }

    const s3Config = JSON.parse(localStorage.getItem('s3config') || '{}');

    const payload = {
        source: source,
        name: backupName,
        bucket: bucketName,
        s3: {
            endpoint: s3Config.endpoint,
            accessKeyId: s3Config.accessKeyId,
            secretKey: s3Config.secretKey,
            useSSL: s3Config.useSSL !== undefined ? s3Config.useSSL : true,
            region: s3Config.region || ''
        }
    };

    messageDiv.innerHTML = '<p>Запуск бэкапа...</p>';

    try {
        const response = await fetch('/api/backup', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (response.ok) {
            const result = await response.json();
            messageDiv.innerHTML = `<p style="color:green;">Бэкап успешно завершён: ${result.message || 'OK'}</p>`;
            // Очищаем поля после успешного бэкапа (опционально)
            document.getElementById('backupSourcePath').value = '';
            document.getElementById('backupName').value = '';
            document.getElementById('backupBucket').value = '';
        } else {
            const errorText = await response.text();
            messageDiv.innerHTML = `<p style="color:red;">Ошибка бэкапа: ${errorText}</p>`;
        }
    } catch (err) {
        messageDiv.innerHTML = `<p style="color:red;">Ошибка соединения: ${err.message}</p>`;
    }
});
let selectedSnapshotName = null;
let selectedSnapshotDate = null;
let selectedSnapshot = null;

// Получаем элементы DOM
const loadSnapshotsBtn = document.getElementById('loadSnapshotsBtn');
const snapshotsListDiv = document.getElementById('snapshotsList');
const restoreBtn = document.getElementById('restoreBtn');
const restoreMessage = document.getElementById('restoreMessage');
const restoreBucket = document.getElementById('restoreBucket');
const restoreTargetPath = document.getElementById('restoreTargetPath');
const snapshotFilesDiv = document.getElementById('snapshotFiles');
const snapshotRestorePath = document.getElementById('snapshotRestorePath');

// Функция для проверки возможности восстановления
function checkRestoreAvailability() {
    const hasSnapshot = selectedSnapshotName !== null;
    const hasTargetPath = restoreTargetPath.value.trim() !== '';
    const hasSnapshotPath = snapshotRestorePath ? snapshotRestorePath.value.trim() !== '' : false;
    
    if (hasSnapshot && hasTargetPath && hasSnapshotPath) {
        restoreBtn.disabled = false;
    } else {
        restoreBtn.disabled = true;
    }
}

// Следим за изменением поля пути восстановления
if (restoreTargetPath) {
    restoreTargetPath.addEventListener('input', () => {
        checkRestoreAvailability();
    });
}

// Следим за изменением поля пути внутри снапшота (опционально)
if (snapshotRestorePath) {
    snapshotRestorePath.addEventListener('input', () => {
        checkRestoreAvailability();
    });
}

// Функция для построения дерева из списка путей
function buildFileTree(paths) {
    const tree = {};
    
    paths.forEach(path => {
        const parts = path.split('/');
        let current = tree;
        
        parts.forEach((part, index) => {
            if (index === parts.length - 1) {
                // Это файл
                if (!current._files) current._files = [];
                current._files.push(part);
            } else {
                // Это папка
                if (!current[part]) current[part] = {};
                current = current[part];
            }
        });
    });
    
    return tree;
}

// Функция для отображения дерева с возможностью выбора пути
function renderTreeWithPath(tree, container, level = 0, parentPath = '') {
    const ul = document.createElement('ul');
    ul.style.listStyle = 'none';
    ul.style.paddingLeft = level === 0 ? '0' : '20px';
    ul.style.margin = '0';
    
    // Сортируем папки
    const folders = Object.keys(tree).filter(key => key !== '_files').sort();
    const files = tree._files ? [...tree._files].sort() : [];
    
    // Отображаем папки
    folders.forEach(folder => {
        const li = document.createElement('li');
        li.style.margin = '8px 0';
        li.style.cursor = 'pointer';
        li.style.userSelect = 'none';
        li.style.padding = '4px 8px';
        li.style.borderRadius = '6px';
        
        const folderIcon = document.createElement('span');
        folderIcon.textContent = '📁 ';
        folderIcon.style.fontSize = '16px';
        
        const folderName = document.createElement('span');
        folderName.textContent = folder;
        folderName.style.fontWeight = '500';
        folderName.style.color = '#e94560';
        folderName.className = 'folder-name';
        
        li.appendChild(folderIcon);
        li.appendChild(folderName);
        
        const childContainer = document.createElement('div');
        childContainer.style.display = 'none';
        childContainer.style.marginTop = '5px';
        
        const currentPath = parentPath ? `${parentPath}/${folder}` : folder;
        
        // Рекурсивно отображаем содержимое папки
        renderTreeWithPath(tree[folder], childContainer, level + 1, currentPath);
        
        li.appendChild(childContainer);
        
        // Обработчик клика по папке
        li.addEventListener('click', (e) => {
            e.stopPropagation();
            if (childContainer.style.display === 'none') {
                childContainer.style.display = 'block';
                folderIcon.textContent = '📂 ';
                // Заполняем поле пути при клике на папку
                if (snapshotRestorePath) {
                    snapshotRestorePath.value = currentPath;
                    checkRestoreAvailability(); // Проверяем после изменения
                }
                // Убираем выделение с других элементов
                document.querySelectorAll('.snapshot-files-list .selected-path').forEach(el => {
                    el.classList.remove('selected-path');
                    el.style.background = 'transparent';
                });
                li.style.background = 'rgba(233, 69, 96, 0.2)';
                li.classList.add('selected-path');
            } else {
                childContainer.style.display = 'none';
                folderIcon.textContent = '📁 ';
            }
        });
        
        ul.appendChild(li);
    });
    
    // Отображаем файлы
    files.forEach(file => {
        const li = document.createElement('li');
        li.style.margin = '5px 0';
        li.style.padding = '4px 8px';
        li.style.borderRadius = '6px';
        li.style.color = '#a0a0d0';
        li.style.fontSize = '13px';
        li.style.fontFamily = 'monospace';
        li.style.cursor = 'pointer';
        
        const fileIcon = document.createElement('span');
        fileIcon.textContent = '📄 ';
        fileIcon.style.fontSize = '14px';
        
        const fileName = document.createElement('span');
        fileName.textContent = file;
        
        li.appendChild(fileIcon);
        li.appendChild(fileName);
        
        const currentPath = parentPath ? `${parentPath}/${file}` : file;
        
        // Обработчик клика по файлу
        li.addEventListener('click', (e) => {
            e.stopPropagation();
            if (snapshotRestorePath) {
                snapshotRestorePath.value = currentPath;
                checkRestoreAvailability(); // Проверяем после изменения
            }
            // Убираем выделение с других элементов
            document.querySelectorAll('.snapshot-files-list .selected-path').forEach(el => {
                el.classList.remove('selected-path');
                el.style.background = 'transparent';
            });
            li.style.background = 'rgba(233, 69, 96, 0.2)';
            li.classList.add('selected-path');
        });
        
        ul.appendChild(li);
    });
    
    container.appendChild(ul);
}

// Загрузка списка снапшотов
loadSnapshotsBtn.addEventListener('click', async () => {
    const bucket = restoreBucket.value.trim();
    if (!bucket) {
        restoreMessage.innerHTML = '<p style="color:red;">Укажите название bucket!</p>';
        return;
    }

    const s3Config = JSON.parse(localStorage.getItem('s3config') || '{}');
    const payload = {
        bucket: bucket,
        s3: {
            endpoint: s3Config.endpoint,
            accessKeyId: s3Config.accessKeyId,
            secretKey: s3Config.secretKey,
            useSSL: s3Config.useSSL !== undefined ? s3Config.useSSL : true,
            region: s3Config.region || ''
        }
    };

    snapshotsListDiv.innerHTML = '<em>Загрузка снапшотов...</em>';

    try {
        const response = await fetch('/api/snapshots', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errorText = await response.text();
            snapshotsListDiv.innerHTML = `<em style="color:red;">Ошибка: ${errorText}</em>`;
            return;
        }
        
        const snapshots = await response.json();
        if (!snapshots.length) {
            snapshotsListDiv.innerHTML = '<em>Нет снапшотов в этом bucket</em>';
            return;
        }

        // Отображение списка снапшотов
        snapshotsListDiv.innerHTML = '';
        snapshots.forEach(snap => {
            const item = document.createElement('div');
            item.className = 'snapshot-item';
            item.textContent = `${snap.name} (${snap.timestamp})`;
            
            item.addEventListener('click', async () => {
    // Убираем выделение со всех
    document.querySelectorAll('.snapshot-item').forEach(el => el.classList.remove('selected'));
    item.classList.add('selected');
    
    // Сохраняем имя и дату снапшота
    selectedSnapshotName = snap.name;
    selectedSnapshotDate = formatSnapshotDate(snap.timestamp);
    
    // Проверяем доступность кнопки после выбора снапшота
    checkRestoreAvailability();
    
    // Очищаем поле пути при выборе нового снапшота
    if (snapshotRestorePath) {
        snapshotRestorePath.value = '';
    }

    // Загружаем список файлов для этого снапшота
    if (snapshotFilesDiv) {
        snapshotFilesDiv.innerHTML = '<em>Загрузка файлов...</em>';
    }

    const bucket = restoreBucket.value.trim();
    const s3Config = JSON.parse(localStorage.getItem('s3config') || '{}');

    const filesPayload = {
        snapshot: snap.name,
        bucket: bucket,
        s3: {
            endpoint: s3Config.endpoint,
            accessKeyId: s3Config.accessKeyId,
            secretKey: s3Config.secretKey,
            useSSL: s3Config.useSSL !== undefined ? s3Config.useSSL : true,
            region: s3Config.region || ''
        }
    };

    try {
        const filesResponse = await fetch('/api/snapshot-files', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(filesPayload)
        });

        if (filesResponse.ok) {
            const files = await filesResponse.json();
            if (files.length === 0) {
                if (snapshotFilesDiv) snapshotFilesDiv.innerHTML = '<em>Нет файлов в этом снапшоте</em>';
            } else {
                // Строим дерево из путей
                const tree = buildFileTree(files);
                
                // Очищаем контейнер и отображаем дерево
                if (snapshotFilesDiv) {
                    snapshotFilesDiv.innerHTML = '';
                    const treeContainer = document.createElement('div');
                    treeContainer.className = 'file-tree';
                    renderTreeWithPath(tree, treeContainer);
                    snapshotFilesDiv.appendChild(treeContainer);
                }
            }
        } else {
            const errorText = await filesResponse.text();
            if (snapshotFilesDiv) snapshotFilesDiv.innerHTML = `<em style="color:red;">Ошибка: ${errorText}</em>`;
        }
    } catch (err) {
        if (snapshotFilesDiv) snapshotFilesDiv.innerHTML = `<em style="color:red;">Ошибка соединения: ${err.message}</em>`;
    }
});
            
            snapshotsListDiv.appendChild(item);
        });
    } catch (err) {
        snapshotsListDiv.innerHTML = `<em style="color:red;">Ошибка соединения: ${err.message}</em>`;
    }
});

function formatSnapshotDate(timestamp) {
    // timestamp имеет формат "2026-05-08T08:44:46Z"
    return timestamp.replace('T', '_').replace('Z', '');
}

// Восстановление выбранного снапшота
restoreBtn.addEventListener('click', async () => {
    if (!selectedSnapshotName) {
        restoreMessage.innerHTML = '<p style="color:red;">Выберите снапшот из списка!</p>';
        return;
    }

    const targetPath = restoreTargetPath.value.trim();
    if (!targetPath) {
        restoreMessage.innerHTML = '<p style="color:red;">Укажите путь для восстановления!</p>';
        return;
    }

    const bucket = restoreBucket.value.trim();
    const s3Config = JSON.parse(localStorage.getItem('s3config') || '{}');
    let snapshotPath = snapshotRestorePath ? snapshotRestorePath.value.trim() : '';
    
    const payload = {
        snapshot: selectedSnapshotName,
        date: selectedSnapshotDate,  // <-- ИСПРАВЛЕНО
        target: targetPath,
        bucket: bucket,
        source: snapshotPath,
        s3: {
            endpoint: s3Config.endpoint,
            accessKeyId: s3Config.accessKeyId,
            secretKey: s3Config.secretKey,
            useSSL: s3Config.useSSL !== undefined ? s3Config.useSSL : true,
            region: s3Config.region || ''
        }
    };
    
    console.log('Payload:', JSON.stringify(payload, null, 2)); // Подробный вывод
    
    restoreMessage.innerHTML = '<p>Запуск восстановления...</p>';

    try {
        const response = await fetch('/api/restore', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        
        if (response.ok) {
            const result = await response.json();
            restoreMessage.innerHTML = `<p style="color:green;">Восстановление завершено: ${result.message || 'OK'}</p>`;
        } else {
            const errorText = await response.text();
            restoreMessage.innerHTML = `<p style="color:red;">Ошибка восстановления: ${errorText}</p>`;
        }
    } catch (err) {
        restoreMessage.innerHTML = `<p style="color:red;">Ошибка соединения: ${err.message}</p>`;
    }
});

// Изначально кнопка заблокирована
restoreBtn.disabled = true;

//стили

const separator = document.querySelector('.separator');
const backupSection = document.querySelector('.backup-section');
const restoreSection = document.querySelector('.restore-section');
let isMoved = false;

separator.addEventListener('click', () => {
    if (!isMoved) {
        separator.classList.add('move-left');
        backupSection.style.transition = 'opacity 0.45s ease';
        backupSection.style.opacity = '0';
        
        setTimeout(() => {
            backupSection.style.display = 'none';
            restoreSection.style.display = 'block';
            restoreSection.style.opacity = '0';
            restoreSection.style.transition = 'opacity 0.45s ease';
            
            setTimeout(() => {
                restoreSection.style.opacity = '1';
            }, 50);
        }, 450);
        
        isMoved = true;
    } else {
        restoreSection.style.opacity = '0';
        
        setTimeout(() => {
            restoreSection.style.display = 'none';
            backupSection.style.display = 'block';
            backupSection.style.opacity = '0';
            
            setTimeout(() => {
                backupSection.style.opacity = '1';
            }, 50);
        }, 450);
        
        separator.classList.remove('move-left');
        isMoved = false;
    }
});