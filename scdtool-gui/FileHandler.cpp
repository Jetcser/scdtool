#include "FileHandler.h"
#include <QFileDialog>
#include <QStandardPaths>
#include <QProcess>
#include <QFile>
#include <QFileInfo>
#include <QTextEdit>
#include <QTextCursor>
#include <QTextCharFormat>
#include <QColor>
#include <QUrl>

QUrl FileHandler::chooseTxtFile(QWidget *parent) {
    QUrl url = QFileDialog::getOpenFileUrl(
        parent,
        "选择文本文件",
        QUrl::fromLocalFile(QStandardPaths::writableLocation(QStandardPaths::DesktopLocation)),
        "文本文件 (*.txt)"
    );
    return url.isEmpty() ? QUrl() : url;
}

QUrl FileHandler::chooseSGTxtFile(QWidget *parent) {
    QUrl url = QFileDialog::getOpenFileUrl(
        parent,
        "选择搜狗文本词库",
        QUrl::fromLocalFile(QStandardPaths::writableLocation(QStandardPaths::DesktopLocation)),
        "搜狗文本词库 (*.txt)"
    );
    return url.isEmpty() ? QUrl() : url;
}

QUrl FileHandler::chooseSCDFile(QWidget *parent) {
    QUrl url = QFileDialog::getOpenFileUrl(
        parent,
        "选择搜狗或QQ细胞词库",
        QUrl::fromLocalFile(QStandardPaths::writableLocation(QStandardPaths::DesktopLocation)),
        "细胞词库 (*.scel *.qcel);;所有文件 (*.*)"
    );
    return url.isEmpty() ? QUrl() : url;
}

void FileHandler::jumpToFile(const QUrl &fileUrl) {
    if (!fileUrl.isLocalFile())
        return;

    QString filePath = fileUrl.toLocalFile();
    if (filePath.isEmpty() || !QFile::exists(filePath))
        return;

    QFileInfo fileInfo(filePath);
    QString dirPath = fileInfo.absolutePath();

    QString ddeFileManager = QStandardPaths::findExecutable("dde-file-manager");
    if (!ddeFileManager.isEmpty()) {
        QProcess::startDetached(ddeFileManager, QStringList() << "--show-item" << filePath);
    } else {
        QString xdgOpen = QStandardPaths::findExecutable("xdg-open");
        if (!xdgOpen.isEmpty()) {
            QProcess::startDetached(xdgOpen, QStringList() << dirPath);
        }
    }
}

std::pair<bool,bool> FileHandler::checkSogouHeader(const QString &filePath) {
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        qDebug() << "无法打开文件:" << filePath;
        return {false, false};
    }

    QByteArray header = file.read(12);  // 读取前12字节
    file.close();

    if (header.size() < 12) {
        qDebug() << "文件太短，无法读取完整文件头";
        return {false, false};
    }

    QByteArray prefix = header.left(4);       // 前4字节
    QByteArray middle = header.mid(4, 2);     // 第5-6字节
    QByteArray suffix = header.mid(6, 6);     // 第7-12字节

    QByteArray expectedPrefix = QByteArray::fromHex("40150000");
    QByteArray expectedSuffix = QByteArray::fromHex("530101000000");  // 6字节

    if (prefix != expectedPrefix || suffix != expectedSuffix) {
        return {false, false};
    }

    bool isOfficial = (middle == QByteArray::fromHex("4443"));
    return {true, isOfficial};
}


// 异步执行 CLI 并输出到 QTextEdit
void FileHandler::runCliTool(const QString &toolPath, const QStringList &arguments, QTextEdit *outputWidget, QWidget *parent) {
    if (toolPath.isEmpty()) {
        outputWidget->append("错误: 工具路径为空");
        return;
    }
    if (!QFile::exists(toolPath)) {
        outputWidget->append("错误: 工具不存在：" + toolPath);
        return;
    }

    auto process = new QProcess(parent);

    auto appendText = [outputWidget](const QString &text, const QColor &color = Qt::black) {
        QTextCharFormat fmt;
        fmt.setForeground(color);
        QTextCursor cursor(outputWidget->document());
        cursor.movePosition(QTextCursor::End);
        cursor.setCharFormat(fmt);
        cursor.insertText(text);
        outputWidget->moveCursor(QTextCursor::End);
    };

    QObject::connect(process, &QProcess::readyReadStandardOutput, [=]() {
        QByteArray data = process->readAllStandardOutput();
        appendText(QString::fromLocal8Bit(data), Qt::black);
    });

    QObject::connect(process, &QProcess::readyReadStandardError, [=]() {
        QByteArray data = process->readAllStandardError();
        appendText(QString::fromLocal8Bit(data), Qt::red);
    });

    QObject::connect(process, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished),
                     [=](int exitCode, QProcess::ExitStatus exitStatus) {
        if (exitStatus == QProcess::NormalExit && exitCode == 0) {
            appendText("\n命令执行完成。\n", Qt::darkGreen);
        } else {
            appendText(QString("\n命令异常退出，退出码：%1\n").arg(exitCode), Qt::red);
        }
        process->deleteLater();
    });

    process->start(toolPath, arguments);
    if (!process->waitForStarted()) {
        appendText("错误: 无法启动命令\n", Qt::red);
        process->deleteLater();
    }
}

// 同步执行 CLI 并返回输出
QString FileHandler::runCliToolSync(const QString &toolPath, const QStringList &arguments) {
    if (toolPath.isEmpty() || !QFile::exists(toolPath)) {
        return QString("错误: 工具不存在或路径为空: %1").arg(toolPath);
    }

    QProcess process;
    process.start(toolPath, arguments);
    if (!process.waitForStarted()) {
        return QString("错误: 无法启动命令 %1").arg(toolPath);
    }

    if (!process.waitForFinished(-1)) {
        return QString("错误: 命令执行超时或异常退出: %1").arg(toolPath);
    }

    QByteArray stdOut = process.readAllStandardOutput();
    QByteArray stdErr = process.readAllStandardError();

    QString output = QString::fromLocal8Bit(stdOut);
    QString error = QString::fromLocal8Bit(stdErr);

    if (!error.isEmpty()) {
        output += "\n[ERROR]: " + error;
    }

    return output;
}
