#include "SCDTool_GUI.h"
#include "ui_SCDTool_GUI.h"
#include "FileHandler.h"
#include <QMessageBox>
#include <QJsonDocument>
#include <QJsonObject>
#include <QCoreApplication>

SCDTool_GUI::SCDTool_GUI(QWidget *parent)
    : QMainWindow(parent), ui(new Ui::SCDTool_GUI) {
    ui->setupUi(this);

    setWindowFlags(windowFlags() & ~Qt::WindowMaximizeButtonHint);
    setFixedSize(this->size());

    ui->infoCheckBoxOfficial->setEnabled(false);
    ui->infoRadioOfficial->setEnabled(false);
    ui->infoRadioOther->setEnabled(false);
    ui->infoResult->setReadOnly(true);

    // 设置 tabTXTMake 为默认显示的 tab
    ui->tabGroup->setCurrentWidget(ui->tabTXTMake);

    setAcceptDrops(true);
    initToolPaths();
}

SCDTool_GUI::~SCDTool_GUI() {
    delete ui;
}

void SCDTool_GUI::initToolPaths() {
    // 获取当前程序所在目录
    QString appDir = QCoreApplication::applicationDirPath();

    // 使用相对路径初始化工具路径
    toolTxtMaker  = appDir + "/txtmaker";
    toolScdMaker  = appDir + "/scdmaker";
    toolScdParser = appDir + "/scdparser";
    toolScdEditor = appDir + "/scdeditor";
}

void SCDTool_GUI::dragEnterEvent(QDragEnterEvent *event) {
    if (event->mimeData()->hasUrls())
        event->acceptProposedAction();
}

void SCDTool_GUI::dropEvent(QDropEvent *event) {
    const auto urls = event->mimeData()->urls();
    if (!urls.isEmpty())
        ui->infoLineChooseFile->setText(urls.first().toLocalFile());
}

// 文本文件按钮
void SCDTool_GUI::on_txtButtonChooseFile_clicked() {
    QUrl fileUrl = FileHandler::chooseTxtFile(this);
    ui->txtLineChooseFile->setText(fileUrl.toLocalFile());
}

void SCDTool_GUI::on_txtButtonConvert_clicked() {
    QString txtFilePath = ui->txtLineChooseFile->text();
    ui->txtResult->clear();
    FileHandler::runCliTool(toolTxtMaker, QStringList() << txtFilePath, ui->txtResult, this);
}

void SCDTool_GUI::on_txtButtonOpenDir_clicked() {
    QUrl fileUrl = QUrl::fromLocalFile(ui->txtLineChooseFile->text());
    FileHandler::jumpToFile(fileUrl);
}

// 搜狗文本按钮
void SCDTool_GUI::on_scdButtonChooseFile_clicked() {
    QUrl fileUrl = FileHandler::chooseSGTxtFile(this);
    ui->scdLineChooseFile->setText(fileUrl.toLocalFile());
}

void SCDTool_GUI::on_scdButtonMake_clicked() {
    QString sgTextFilePath = ui->scdLineChooseFile->text();
    ui->scdResult->clear();
    FileHandler::runCliTool(toolScdMaker, QStringList() << sgTextFilePath, ui->scdResult, this);
}

void SCDTool_GUI::on_scdButtonOpenDir_clicked() {
    QString sgTextFilePath = ui->scdLineChooseFile->text();
    sgTextFilePath.replace(".txt", ".scel");
    QUrl fileUrl = QUrl::fromLocalFile(sgTextFilePath);
    FileHandler::jumpToFile(fileUrl);
}

// 细胞词库按钮
void SCDTool_GUI::on_infoButtonChooseFile_clicked() {
    QUrl fileUrl = FileHandler::chooseSCDFile(this);
    ui->infoLineChooseFile->setText(fileUrl.toLocalFile());
}

void SCDTool_GUI::on_infoButtonParse_clicked() {
    QString scelPath = ui->infoLineChooseFile->text();
    if (scelPath.isEmpty()) {
        QMessageBox::warning(this, tr("警告"), tr("请选择词库文件"));
        return;
    }

    // 使用新函数，返回 pair<bool,bool>
    auto [isValidFile, isOfficialFile] = FileHandler::checkSogouHeader(scelPath);

    if (!isValidFile) {
        QMessageBox::warning(this, tr("错误"), tr("文件头不正确，这似乎不是有效的搜狗细胞词库文件"));
        return;
    }

    ui->infoResult->clear();
    ui->infoResult->append(tr("文件头检测通过，开始解析词库..."));

    // Radiobutton 只读
    ui->infoRadioOfficial->setChecked(isOfficialFile);
    ui->infoRadioOfficial->setEnabled(false);
    ui->infoRadioOther->setChecked(!isOfficialFile);
    ui->infoRadioOther->setEnabled(false);

    // Checkbox
    if (isOfficialFile) {
        ui->infoCheckBoxOfficial->setChecked(false);  // 不勾选
        ui->infoCheckBoxOfficial->setEnabled(false);  // 禁用
    } else {
        ui->infoCheckBoxOfficial->setChecked(false);  // 默认不勾选
        ui->infoCheckBoxOfficial->setEnabled(true);   // 可编辑
    }

    // 执行 CLI 工具同步解析
    QString output = FileHandler::runCliToolSync(toolScdParser, QStringList() << scelPath);

    QJsonParseError jsonError;
    QJsonDocument doc = QJsonDocument::fromJson(output.toUtf8(), &jsonError);
    if (jsonError.error != QJsonParseError::NoError) {
        ui->infoResult->append(tr("解析失败: %1").arg(jsonError.errorString()));
        return;
    }

    QJsonObject obj = doc.object();

    // 填充 UI
    ui->infoLineDictId->setText(obj.value("id").toString());
    ui->infoLineDictName->setText(obj.value("name").toString());
    ui->infoLineCategory->setText(obj.value("category").toString());
    ui->infoLineDictRemark->setText(obj.value("remark").toString());

    // 保存原始值
    originalDictId = ui->infoLineDictId->text();
    originalDictName = ui->infoLineDictName->text();
    originalCategory = ui->infoLineCategory->text();
    originalRemark = ui->infoLineDictRemark->text();
    originalOfficial = ui->infoCheckBoxOfficial->isChecked();

    ui->infoResult->append(tr("解析完成"));
}



// 修改词库信息
void SCDTool_GUI::on_infoButtonModify_clicked() {
    QString scdFilePath = ui->infoLineChooseFile->text();
    if (scdFilePath.isEmpty()) {
        QMessageBox::warning(this, tr("警告"), tr("请选择词库文件"));
        return;
    }

    // 检查是否有修改
    if (ui->infoLineDictId->text() == originalDictId &&
        ui->infoLineDictName->text() == originalDictName &&
        ui->infoLineCategory->text() == originalCategory &&
        ui->infoLineDictRemark->text() == originalRemark &&
        ui->infoCheckBoxOfficial->isChecked() == originalOfficial) {
        ui->infoResult->append(tr("信息未修改，跳过操作"));
        return;
        }

    ui->infoResult->clear();
    ui->infoResult->append(tr("开始运行 scdeditor 修改词库信息..."));

    QStringList args;
    args << "-f" << scdFilePath;  // 必须指定文件路径
    if (!ui->infoLineDictId->text().isEmpty())       args << "-i" << ui->infoLineDictId->text();
    if (!ui->infoLineDictName->text().isEmpty())     args << "-n" << ui->infoLineDictName->text();
    if (!ui->infoLineCategory->text().isEmpty())     args << "-c" << ui->infoLineCategory->text();
    if (!ui->infoLineDictRemark->text().isEmpty())   args << "-r" << ui->infoLineDictRemark->text();

    // 仅当 checkbox 选中时，才修改词头
    if (ui->infoCheckBoxOfficial->isChecked()) {
        args << "-o";  // 标记为官方词库
    }

    FileHandler::runCliTool(toolScdEditor, args, ui->infoResult, this);

    // 修改完成后更新原始值
    originalDictId = ui->infoLineDictId->text();
    originalDictName = ui->infoLineDictName->text();
    originalCategory = ui->infoLineCategory->text();
    originalRemark = ui->infoLineDictRemark->text();
    originalOfficial = ui->infoCheckBoxOfficial->isChecked();
}
