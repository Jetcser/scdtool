#ifndef SCDTOOL_GUI_H
#define SCDTOOL_GUI_H

#include <QMainWindow>
#include <QLineEdit>
#include <QFileDialog>
#include <QTextEdit>
#include <QPushButton>
#include <QCheckBox>
#include <QRadioButton>
#include <QDropEvent>
#include <QMimeData>
#include <QDragEnterEvent>
#include <QUrl>
#include <QByteArray>
#include <QFile>

namespace Ui {
    class SCDTool_GUI;
}

// 文件头三段结构
struct FileHeaderSegment {
    QByteArray prefix;   // 前段固定字节
    QByteArray variable; // 中段可变字节
    QByteArray suffix;   // 后段固定字节
};

class SCDTool_GUI : public QMainWindow {
    Q_OBJECT

public:
    explicit SCDTool_GUI(QWidget *parent = nullptr);
    ~SCDTool_GUI() override;

protected:
    void dragEnterEvent(QDragEnterEvent *event) override;
    void dropEvent(QDropEvent *event) override;

private slots:
    void on_txtButtonChooseFile_clicked();
    void on_txtButtonConvert_clicked();
    void on_txtButtonOpenDir_clicked();

    void on_scdButtonChooseFile_clicked();
    void on_scdButtonMake_clicked();
    void on_scdButtonOpenDir_clicked();

    void on_infoButtonChooseFile_clicked();
    void on_infoButtonParse_clicked();
    void on_infoButtonModify_clicked();

private:
    Ui::SCDTool_GUI *ui{nullptr};

    // CLI 工具路径
    QString toolTxtMaker;
    QString toolScdMaker;
    QString toolScdParser;
    QString toolScdEditor;

    // 原始词库信息，用于检测是否修改
    QString originalDictId;
    QString originalDictName;
    QString originalCategory;
    QString originalRemark;
    bool originalOfficial = false;

    void initToolPaths(); // 初始化工具路径

    // 文件头三段常量（官方词库）
    static inline const FileHeaderSegment OFFICIAL_HEADER_SEGMENT = {
        QByteArray::fromHex("40150000"),   // 前4字节固定
        QByteArray::fromHex("4443"),       // 中段可变
        QByteArray::fromHex("5301010000")  // 后6字节固定
    };

    // 判断文件是否为合法搜狗细胞词库
    static bool isValidSogouHeader(const QString &filePath) {
        QFile file(filePath);
        if (!file.open(QIODevice::ReadOnly)) return false;
        QByteArray header = file.read(12);
        file.close();
        if (header.size() < 12) return false;

        return (header.left(4) == OFFICIAL_HEADER_SEGMENT.prefix &&
                header.right(6) == OFFICIAL_HEADER_SEGMENT.suffix);
    }

    // 判断文件头是否官方（中段可变，不比对中段）
    static bool isOfficialHeader(const QByteArray &header) {
        if (header.size() < 12) return false;
        return (header.left(4) == OFFICIAL_HEADER_SEGMENT.prefix &&
                header.right(6) == OFFICIAL_HEADER_SEGMENT.suffix);
    }
};

#endif // SCDTOOL_GUI_H
