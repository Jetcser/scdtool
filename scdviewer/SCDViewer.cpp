#include "SCDInfoRead.h"
#include <QApplication>
#include <QWidget>
#include <QLabel>
#include <QVBoxLayout>
#include <QPushButton>
#include <QMessageBox>
#include <QFileInfo>

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);

    if (argc < 2) {
        QMessageBox::critical(nullptr, "错误", "请提供词库文件路径");
        return 1;
    }

    QString filePath = argv[1];
    QFileInfo fileInfo(filePath);

    if (!fileInfo.exists() || !fileInfo.isFile()) {
        QMessageBox::critical(nullptr, "错误", "文件不存在或不是有效文件:\n" + filePath);
        return 1;
    }

    SCDInfo info = SCDInfoRead::readSCDInfo(filePath);

    if (info.allInformation.startsWith("非法")) {
        QMessageBox::critical(nullptr, "错误", info.allInformation);
        return 1;
    }

    // 创建主窗口
    QWidget window;
    window.setWindowTitle("细胞词库信息");

    // 创建主布局（垂直）
    QVBoxLayout *layout = new QVBoxLayout(&window);

    // === 文本显示区域 ===
    QWidget *textContainer = new QWidget(&window);
    QVBoxLayout *textLayout = new QVBoxLayout(textContainer);

    // 创建文本标签
    QLabel *label = new QLabel();
    label->setText(info.allInformation);
    label->setWordWrap(true);
    label->setTextInteractionFlags(Qt::TextSelectableByMouse);

    // 设置字号
    QFont font = label->font();  // 获取默认字体
    font.setPointSize(10);
    label->setFont(font);

    // 将标签加入子布局
    textLayout->addWidget(label);

    // 根据窗口大小计算百分比内边距（仅作用于文本部分）
    textContainer->adjustSize();
    QSize textSize = textContainer->size();
    int paddingW = static_cast<int>(textSize.width() * 0.1);
    int paddingH = static_cast<int>(textSize.height() * 0.1);
    textLayout->setContentsMargins(paddingW, paddingH, paddingW, paddingH);

    // 将带边距的文本区域加入主布局
    layout->addWidget(textContainer);

    // === 确认按钮 ===
    QPushButton *okButton = new QPushButton("确认");
    okButton->setFixedWidth(80);
    okButton->setCursor(Qt::PointingHandCursor);
    layout->addWidget(okButton, 0, Qt::AlignCenter);

    // 信号连接：点击确认 → 退出程序
    QObject::connect(okButton, &QPushButton::clicked, &app, &QApplication::quit);

    // 调整窗口大小
    window.adjustSize();
    window.setMaximumSize(1000, 800);
    window.show();

    return app.exec();
}
