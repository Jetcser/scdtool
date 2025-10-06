#ifndef FILEHANDLER_H
#define FILEHANDLER_H

#include <QWidget>
#include <QTextEdit>
#include <QStringList>
#include <QUrl>

namespace FileHandler {

    // 选择中文词条文本文件
    QUrl chooseTxtFile(QWidget *parent = nullptr);

    // 选择搜狗文本词库文件
    QUrl chooseSGTxtFile(QWidget *parent = nullptr);

    // 选择搜狗或QQ细胞词库文件
    QUrl chooseSCDFile(QWidget *parent = nullptr);

    // 跳转到文件所在目录（QUrl 或 QString 均可）
    void jumpToFile(const QUrl &fileUrl);

    // 验证搜狗词库文件头
    // 返回 pair：{是否有效词库, 是否官方词库}
    std::pair<bool, bool> checkSogouHeader(const QString &filePath);

    // 异步运行 CLI 工具并输出到 QTextEdit
    void runCliTool(const QString &toolPath,
                    const QStringList &arguments,
                    QTextEdit *outputWidget,
                    QWidget *parent = nullptr);

    // 同步运行 CLI 工具，返回输出字符串
    QString runCliToolSync(const QString &toolPath,
                           const QStringList &arguments);

}

#endif // FILEHANDLER_H
